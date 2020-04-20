package handler

import (
	"fmt"

	"github.com/ronenlib/kube-failure-alert/pkg/notifier"
	corev1 "k8s.io/api/core/v1"
)

// Handler handles controller events
type Handler interface {
	Handle(obj interface{}) error
}

// EventHandler handels incoming events and notifies when
// an event is alerting
type EventHandler struct{
	notifier notifier.Notifier
}

// NewEventHandler creates new event handler
func NewEventHandler(notifier notifier.Notifier) *EventHandler {
	return &EventHandler{
		notifier: notifier,
	}
}

// Handle checks event health and notifies if unhealthy
func (h *EventHandler) Handle(obj interface{}) error {
	event, ok := obj.(*corev1.Event)

	if !ok {
		return fmt.Errorf("handler expects event object, got %t", obj)
	}

	if event.Type == corev1.EventTypeNormal {
		return nil
	}

	payload := h.getPayload(event)

	return h.notifier.Notify(payload)
}

func (h *EventHandler) getPayload(event *corev1.Event) notifier.Payload {
	return notifier.Payload{
		Kind:      event.InvolvedObject.Kind,
		Namespace: event.InvolvedObject.Namespace,
		Name:      event.InvolvedObject.Name,
		Error:     fmt.Sprintf("%s %s", event.Reason, event.Message),
	}
}
