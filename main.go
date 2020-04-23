package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ronenlib/kube-event-alert/pkg/controller"
	"github.com/ronenlib/kube-event-alert/pkg/notifier"
	"github.com/ronenlib/kube-event-alert/pkg/util"
	"k8s.io/klog"
)

var (
	kubeconfig string
	masterURL  string
	webhookURL string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	clientset := util.GetKubeClient(masterURL, kubeconfig)
	// todo remvoe url
	notifier := notifier.NewSlackNotifier(webhookURL)

	stopCh := make(chan struct{})
	setInterrupt(stopCh)

	controller.Launch(clientset, notifier, stopCh)
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
	flag.StringVar(&webhookURL, "webhookURL", "", "notification will be sent to this slack incoming webhook url")
}
