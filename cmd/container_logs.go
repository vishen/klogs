package cmd

import (
	"context"
	"os"

	"github.com/vishen/go-slearch/slearch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ContainerLogs refers to a container in a kubernetes cluster
type ContainerLogs struct {
	name      string
	podName   string
	namespace string

	tailLogs bool

	slearchConfig slearch.Config
	client        *kubernetes.Clientset
}

// NewContainerLOgs returns a new container logs instance
func NewContainerLogs(name, podName, namespace string, client *kubernetes.Clientset, slearchConfig slearch.Config) *ContainerLogs {
	return &ContainerLogs{
		name:          name,
		podName:       podName,
		namespace:     namespace,
		client:        client,
		slearchConfig: slearchConfig,
		tailLogs:      true,
	}
}

// Start takes a ctx and will start searching on a containers log strea,
func (cl ContainerLogs) Start(ctx context.Context) {
	cl.searchLogs(ctx)
}

// searchLogs will connect to the kubernetes container stream and pass the
// stream to 'slearch'. 'slearch' will search the log stream based on the config
func (cl ContainerLogs) searchLogs(ctx context.Context) {

	r := cl.client.CoreV1().Pods(cl.namespace).GetLogs(cl.podName, &corev1.PodLogOptions{
		Container: cl.name,
		Follow:    cl.tailLogs,
	}).Context(ctx)

	stream, err := r.Stream()
	if err != nil {
		// log.Printf("%s, %s, %s: cannot get stream: %s\n", cl.namespace, cl.podName, cl.name, err)
		return
	}
	defer stream.Close()

	if err := slearch.StructuredLoggingSearch(cl.slearchConfig, stream, os.Stdout); err != nil {
		// log.Printf("error searching structured logs: %s\n", err)
	}
}
