package config

import (
	"flag"
	"log"
	"os"
)

// Config for kube event alert controller
type Config struct {
	Kubeconfig string
	MasterURL  string
	WebhookURL string
}

// Load configuration
func Load() Config {
	c := Config{
		MasterURL:  os.Getenv("KUBE_MASTER_URL"),
		WebhookURL: os.Getenv("WEBHOOK_URL"),
	}

	flag.StringVar(&c.Kubeconfig, "kubeconfig", "", "path to a kubeconfig if running out of cluster")
	flag.StringVar(&c.MasterURL, "masterUrl", c.MasterURL,
		"url to kube cluster if running out of cluster"+
			"Can be specified with KUBE_MASTER_URL environment variable as well")
	flag.StringVar(&c.WebhookURL, "webhookURL", c.WebhookURL,
		"notification will be sent to this webhook url."+
			"Can be specified with WEBHOOK_URL environment variable as well")
	flag.Parse()

	if c.WebhookURL == "" {
		log.Fatal("Expected WEBHOOK_URL environment variable or webhookURL argument to be set")
	}

	return c
}
