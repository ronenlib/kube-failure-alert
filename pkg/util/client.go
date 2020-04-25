package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// GetKubeClient returns k8s clientset. If called out of cluster either masterURL
// or kubeconfig should be specified
func GetKubeClient(masterURL, kubeconfig string) kubernetes.Interface {
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("failed to build config: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to build clientset: %s", err.Error())
	}

	return clientset
}
