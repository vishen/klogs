package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	// Required for gcp authentication to gke
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"

	// Load slearch formatters: json and text
	_ "github.com/vishen/go-slearch/formatters"
	"github.com/vishen/go-slearch/slearch"
)

func logs(cmd *cobra.Command, args []string) {
	slearchConfig := getSlearchConfig(cmd, args)

	kubeConfig, _ := cmd.Flags().GetString("kubeconfig")
	kubeContext, _ := cmd.Flags().GetString("kubecontext")
	namespace, _ := cmd.Flags().GetString("namespace")
	selector, _ := cmd.Flags().GetString("selector")
	containers, _ := cmd.Flags().GetStringSlice("containers")

	// TODO(vishen): add resource name arguments similar to kubectl
	// TODO(vishen): add cmd arg for tailing or not
	// TODO(vishen): watch for new pods
	// TODO(vishen): option to not close streams gracefully?

	// Determine kubeconfig path
	if kubeConfig == "" {
		if os.Getenv("KUBECONFIG") != "" {
			kubeConfig = os.Getenv("KUBECONFIG")
		} else {
			kubeConfig = clientcmd.RecommendedHomeFile
		}
	}

	// Create the kubernetes client configuration
	clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: kubeConfig,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: kubeContext,
		},
	).ClientConfig()
	if err != nil {
		log.Fatalf("unable to create k8s client config: %s\n", err)
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Fatalf("unable to create k8s client: %s\n", err)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}

	// Create a new watcher to monitor log streams
	watcher := NewContainerLogsWatcher()

	// Get all pods in a cluster
	pods, err := client.CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		log.Fatalf("unable to get pods: %s", err)
	}

	for _, p := range pods.Items {

		podName := p.Name
		podNamespace := p.Namespace
		podContainers := p.Spec.Containers

		for _, pc := range podContainers {
			log.Printf("%s, %s, %s attached\n", podNamespace, podName, pc.Name)

			if len(containers) > 0 {
				// Check to see if there is this is a wanted container name
				valid := false
				for _, c := range containers {
					if pc.Name == c {
						valid = true
						break
					}
				}
				if !valid {
					break
				}
			}

			// Add the pod / container specifc keys
			slearchConfig.Extras = []slearch.KV{
				slearch.KV{Key: "pod_name", Value: podName},
				slearch.KV{Key: "namespace", Value: podNamespace},
				slearch.KV{Key: "container_name", Value: pc.Name},
			}

			// Add a prefix for when we aren't trying to parse the line
			slearchConfig.Prefix = fmt.Sprintf("[%s] %s (%s) - ", podNamespace, podName, pc.Name)

			// Add the container to the watcher to monitor and search the log streams
			watcher.AddContainer(pc.Name, podName, podNamespace, client, slearchConfig)
		}
	}

	// Catch any interrupt signals and gracefully close everything
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Wait for any signals and stop the watcher
	go func() {
		select {
		case <-sigCh:
			fmt.Println("gracefully closing all log streams")
			watcher.Stop()
		}
	}()

	// Wait for all the container log streams to finish
	<-watcher.DoneChan()
	fmt.Println("finished closing streams")

}
