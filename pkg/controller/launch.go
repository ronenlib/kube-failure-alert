package controller

import (
	"github.com/ronenlib/kube-failure-alert/pkg/handler"
	"github.com/ronenlib/kube-failure-alert/pkg/notifier"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// todo remove fake notifier
type fakeNotifier struct{}

func (f fakeNotifier) Notify(payload notifier.Payload) error {
	return nil
}

// Launch creates and executes controller
func Launch(clientset kubernetes.Interface, stopCh <-chan struct{}) {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events()

	eventHandler := handler.NewEventHandler(fakeNotifier{})

	c := newController("event", clientset, informer, eventHandler)

	klog.Info("start informers")
	factory.Start(stopCh)

	klog.Info("wait for informers cache sync...")
	factory.WaitForCacheSync(stopCh)
	klog.Info("informers cache synced")

	c.Run(stopCh)
}
