package cmd

import (
	"context"
	"sync"

	"github.com/vishen/go-slearch/slearch"
	"k8s.io/client-go/kubernetes"
)

// ContainerLogsWatcheris a watcher to monitor and control all container log streams
type ContainerLogsWatcher struct {
	wg sync.WaitGroup

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
}

// NewContainerLogsWatcher returns a new watcher
func NewContainerLogsWatcher() *ContainerLogsWatcher {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	return &ContainerLogsWatcher{wg: wg, ctx: ctx, ctxCancelFunc: cancel}
}

// AddContainer adds a container to the watcher and starts searching the container
// log stream
func (w *ContainerLogsWatcher) AddContainer(name, podName, namespace string, client *kubernetes.Clientset, slearchConfig slearch.Config) {
	cl := NewContainerLogs(name, podName, namespace, client, slearchConfig)

	w.wg.Add(1)
	go func() {
		cl.Start(w.ctx)
		w.wg.Done()
	}()
}

// Stop will cancel the ctx for each container log stream
func (w *ContainerLogsWatcher) Stop() {
	w.ctxCancelFunc()
}

// DoneChan returns a channel that indicates whether the all the container log
// streams have finished
func (w *ContainerLogsWatcher) DoneChan() chan struct{} {
	finishedChan := make(chan struct{}, 1)

	go func() {
		w.wg.Wait()
		close(finishedChan)
	}()

	return finishedChan
}
