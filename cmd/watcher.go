package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	// Load slearch formatters: json and text
	_ "github.com/vishen/go-slearch/formatters"
	"github.com/vishen/go-slearch/slearch"
)

// WatcherConfig holds information about a pod selections
type WatcherConfig struct {
	namespace string
	selector  string
	tail      bool

	// Pod and container names to filter on
	validPodNames       []string
	validContainerNames []string

	// Slearch config to use
	slearchConfig slearch.Config
}

// ContainerLogsWatcheri is a watcher to monitor and control all container log streams
type ContainerLogsWatcher struct {
	wg           sync.WaitGroup
	finishedChan chan struct{}

	client *kubernetes.Clientset
	config WatcherConfig

	mu               sync.Mutex
	podsBeingWatched map[string]bool
}

// NewContainerLogsWatcher returns a new watcher
func NewContainerLogsWatcher(client *kubernetes.Clientset, config WatcherConfig) *ContainerLogsWatcher {
	wg := sync.WaitGroup{}
	return &ContainerLogsWatcher{
		wg:               wg,
		client:           client,
		config:           config,
		mu:               sync.Mutex{},
		podsBeingWatched: make(map[string]bool),
	}
}

// Start will get the selected pods and watch their containers
// log streams
func (w *ContainerLogsWatcher) Start(ctx context.Context) {

	listOptions := metav1.ListOptions{
		LabelSelector: w.config.selector,
	}

	// Get all pods in a cluster
	pods, err := w.client.CoreV1().Pods(w.config.namespace).List(listOptions)
	if err != nil {
		log.Fatalf("unable to get pods: %s", err)
	}

	for _, p := range pods.Items {
		w.AddPod(ctx, p)
	}

	// If we are not tailing the logs, bail out
	if !w.config.tail {
		return
	}

	// Start the watcher function that will watch kubernetes pod events
	// and start watching the new pods containers log streams
	go func() {
		listOptions.Watch = true
		podWatcher, err := w.client.CoreV1().Pods(w.config.namespace).Watch(listOptions)
		if err != nil {
			log.Printf("unable to watch for new pods: %s", err)
		}

		for {
			select {
			case event := <-podWatcher.ResultChan():
				if event.Type != watch.Added && event.Type != watch.Modified {
					continue
				}

				pod := event.Object.(*corev1.Pod)
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}

				w.AddPod(ctx, *pod)
			case <-ctx.Done():
				podWatcher.Stop()
				return
			}
		}

	}()

}

// AddPod will start a log stream for each of the pods containers
func (w *ContainerLogsWatcher) AddPod(ctx context.Context, pod corev1.Pod) {
	name := pod.Name

	// Check if we are already watching this pod
	w.mu.Lock()
	if _, ok := w.podsBeingWatched[name]; ok {
		w.mu.Unlock()
		return
	}
	w.podsBeingWatched[name] = true
	w.mu.Unlock()

	// Check to see if the pod name is apart of the valid pod names,
	// if we have them
	if len(w.config.validPodNames) > 0 {
		valid := false
		for _, n := range w.config.validPodNames {
			if name == n {
				valid = true
			}
		}
		if !valid {
			return
		}
	}

	slearchConfig := w.config.slearchConfig

	namespace := pod.Namespace
	containers := pod.Spec.Containers

	for _, container := range containers {

		if len(w.config.validContainerNames) > 0 {
			// Check to see if there is this is a wanted container name
			valid := false
			for _, c := range w.config.validContainerNames {
				if container.Name == c {
					valid = true
					break
				}
			}
			if !valid {
				break
			}
		}

		w.wg.Add(1)
		go func(containerName string, slearchConfig slearch.Config) {
			defer w.wg.Done()
			log.Printf("%s, %s, %s attached\n", namespace, name, containerName)

			stream, err := w.client.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
				Container: containerName,
				Follow:    w.config.tail,
			}).Context(ctx).Stream()
			if err != nil {
				log.Printf("error connecting to pod '%s' container '%s': %s\n", name, containerName, err)
				return
			}
			defer stream.Close()

			if len(containers) > 1 {
				// Add the pod / container specifc keys
				slearchConfig.Extras = []slearch.KV{
					slearch.KV{Key: "pod_name", Value: name},
					slearch.KV{Key: "namespace", Value: namespace},
					slearch.KV{Key: "container_name", Value: containerName},
				}

				// Add a prefix for when we aren't trying to parse the line
				slearchConfig.Prefix = fmt.Sprintf("[%s] %s (%s): ", namespace, name, containerName)
			} else {
				// If there is only one container, leave out the container name
				// TODO(vishen): Does this make sense?

				// Add the pod / container specifc keys
				slearchConfig.Extras = []slearch.KV{
					slearch.KV{Key: "pod_name", Value: name},
					slearch.KV{Key: "namespace", Value: namespace},
				}

				// Add a prefix for when we aren't trying to parse the line
				slearchConfig.Prefix = fmt.Sprintf("[%s] %s: ", namespace, name)
			}

			if err := slearch.StructuredLoggingSearch(slearchConfig, stream, os.Stdout); err != nil {
				log.Printf("finished slearch with errors: %s\n", err)
			}

		}(container.Name, slearchConfig)

	}
}

// DoneChan returns a channel that indicates whether the all the container log
// streams have finished
func (w *ContainerLogsWatcher) DoneChan() chan struct{} {
	w.finishedChan = make(chan struct{}, 1)

	go func() {
		w.wg.Wait()
		log.Printf("finished waiting\n")
		close(w.finishedChan)
	}()

	return w.finishedChan
}

// ForceFinish will force close the finished chan if there is something
// blocking for any reason
func (w *ContainerLogsWatcher) ForceFinish() {
	log.Println("Force finishing")
	close(w.finishedChan)
}
