package controller

import (
	"github.com/ronenlib/kube-event-alert/pkg/handler"
	"github.com/ronenlib/kube-event-alert/pkg/notifier"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Launch creates and executes controller
func Launch(clientset kubernetes.Interface, notifier notifier.SlackNotifier, stopCh <-chan struct{}) {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events()

	eventHandler := handler.NewEventHandler(notifier)

	c := newController("event", clientset, informer, eventHandler)

	klog.Info("start informers")
	factory.Start(stopCh)

	klog.Info("wait for informers cache sync...")
	factory.WaitForCacheSync(stopCh)
	klog.Info("informers cache synced")

	c.Run(stopCh)
}
