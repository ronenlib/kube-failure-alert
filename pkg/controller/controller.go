package controller

import (
	"fmt"
	"time"

	"github.com/ronenlib/kube-event-alert/pkg/handler"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// Controller implement the logic of kube object notifier
type Controller struct {
	name          string
	kubeClientset kubernetes.Interface
	informer      cache.SharedIndexInformer
	lister        listers.EventLister
	workqueue     workqueue.RateLimitingInterface
	handler       handler.Handler
}

func newController(name string, clientset kubernetes.Interface, informer informers.EventInformer, handler handler.Handler) *Controller {
	defaultQueue := workqueue.DefaultControllerRateLimiter()
	queueName := fmt.Sprintf("kube-event-alert-%s", name)
	queue := workqueue.NewNamedRateLimitingQueue(defaultQueue, queueName)

	c := &Controller{
		name:          name,
		kubeClientset: clientset,
		informer:      informer.Informer(),
		lister:        informer.Lister(),
		workqueue:     queue,
		handler:       handler,
	}

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueEvent(obj)
		},
	})

	return c
}

// Run controller worker which will handle events
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting worker...")
	go wait.Until(c.runWorker, time.Second, stopCh)
	klog.Info("worker started")

	<-stopCh
	klog.Info("stop worker")
}

func (c *Controller) runWorker() {
	var continueProcess bool
	var err error

	for continueProcess = true; continueProcess; continueProcess, err = c.processNextWorkItem() {
		if err != nil {
			runtime.HandleError(err)
		}
	}
}

func (c *Controller) processNextWorkItem() (bool, error) {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false, nil
	}

	defer c.workqueue.Done(obj)

	key, ok := obj.(string)

	if !ok {
		c.workqueue.Forget(obj)
		return true, fmt.Errorf("unknown type received by workqueue %#v", obj)
	}

	err := c.handleKey(key)

	if err == nil {
		klog.Infof("Successfully handeled %s", key)
	} else {
		klog.Infof("Failed to handel %s, sending back to queue", key)
		c.workqueue.AddRateLimited(obj)
	}

	c.workqueue.Forget(obj)
	return true, err
}

func (c *Controller) handleKey(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		return err
	}

	obj, err := c.lister.Events(namespace).Get(name)

	if err != nil {
		return err
	}

	return c.handler.Handle(obj)
}

func (c *Controller) enqueueEvent(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.workqueue.Add(key)
}
