package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ronenlib/kube-failure-alert/pkg/controller"
	"github.com/ronenlib/kube-failure-alert/pkg/util"
	"k8s.io/klog"
)

var (
	kubeconfig string
	masterURL  string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	clientset := util.GetKubeClient(masterURL, kubeconfig)

	stopCh := make(chan struct{})
	setInterrupt(stopCh)

	controller.Launch(clientset, stopCh)
}

func setInterrupt(stopCh chan struct{}) {
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		close(stopCh)

		<-interrupt
		os.Exit(1)
	}()
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to a kubeconfig if running out of cluster")
	flag.StringVar(&masterURL, "masterUrl", "", "url to kube cluster if running out of cluster")
}
