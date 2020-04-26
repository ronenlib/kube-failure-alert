# KUBE EVENT ALERT

Kube event alert is a k8s controller that watches for events in the cluster and alerts about any abnormal events by calling a webhook with the event details. The webhook payload support slack incoming webhook payload. The controller can run both in and out of the k8s cluster.

## Usage

```
go run main.go [OPTIONS]

OPTIONS:
-kubeconfig
    path to a kubeconfig if running out of cluster
-masterUrl
    url to kube cluster if running out of cluster
    can be specified with KUBE_MASTER_URL environment variable as well
-webhookURL
    notification will be sent to this slack incoming webhook url
    can be specified with WEBHOOK_URL environment variable as well
```

## Installation

### In Cluster
Copy config map yaml file, set the webhook url `webhook.url`

```bash
cp configmap.example.yaml configmap.yaml
```

Apply config map, RBAC authorization and the controller pod to your k8s cluster

```
kubectl apply -f role.yaml
kubectl apply -f configmap.yaml
kubectl apply -f kube-event-alert.yaml
```

### Out of Cluster

#### Executable
Compile an executable file

```
go mod download
go build -o kube-event-alert .
```

Run the executable according to the [usage](#usage) instructions above and satisfy `webhookURL` argument and either the `kubeconfig` or `masterUrl` arguments

#### Docker
```
docker build -t <name>:<tag> .
docker run -d -e WEBHOOK_URL=<webhook-url> -e KUBE_MASTER_URL=<kube-master-url> <name>:<tag>
```

## Notification
The event notification is sent to the webhook url via HTTP `POST` method with the following payload

```json
{
    "text": "event details"
}
```

The payload support slack incoming webhook integration