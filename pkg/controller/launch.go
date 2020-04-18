package controller

import (
	"github.com/ronenlib/kube-failure-alert/pkg/handler"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Launch creates and executes controller
func Launch(clientset kubernetes.Interface, stopCh <-chan struct{}) {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Pods()

	tempHandler := handler.TmpHandler{}

	c := newController("pod", clientset, informer, tempHandler)

	klog.Info("start informers")
	factory.Start(stopCh)

	klog.Info("wait for informers cache sync...")
	factory.WaitForCacheSync(stopCh)
	klog.Info("informers cache synced")

	c.Run(stopCh)
}
