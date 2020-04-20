package handler

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ronenlib/kube-failure-alert/pkg/notifier"
	corev1 "k8s.io/api/core/v1"
)

type mockNotifier struct {
	withError    bool
	lastReceived notifier.Payload
}

func newMockNotifier(withError bool) *mockNotifier {
	return &mockNotifier{
		withError: withError,
	}
}

func (s *mockNotifier) Notify(payload notifier.Payload) error {
	s.lastReceived = payload

	if s.withError {
		return errors.New("failed to notify")
	}

	return nil
}

func newEvent(eventType string) *corev1.Event {
	return &corev1.Event{
		InvolvedObject: corev1.ObjectReference{
			Kind:      "kind",
			Namespace: "namespace",
			Name:      "name",
		},
		Type:    eventType,
		Reason:  "reason",
		Message: "message",
	}
}

func TestHandle(t *testing.T) {
	cases := []struct {
		notifier          *mockNotifier
		arg               interface{}
		notifierTriggered bool
		expectError       bool
		name              string
	}{
		{
			notifier:          newMockNotifier(false),
			arg:               newEvent(corev1.EventTypeNormal),
			notifierTriggered: false,
			expectError:       false,
			name:              "NormalEvent",
		},
		{
			notifier:          newMockNotifier(false),
			arg:               newEvent(corev1.EventTypeWarning),
			notifierTriggered: true,
			expectError:       false,
			name:              "WarningEvent",
		},
		{
			notifier:          newMockNotifier(false),
			arg:               struct{}{},
			notifierTriggered: false,
			expectError:       true,
			name:              "BadType",
		},
		{
			notifier:          newMockNotifier(true),
			arg:               newEvent(corev1.EventTypeWarning),
			notifierTriggered: true,
			expectError:       true,
			name:              "ErrorNotifier",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewEventHandler(tc.notifier)
			err := h.Handle(tc.arg)

			if err != nil && !tc.expectError {
				t.Error("Expected error")
			}

			if tc.notifierTriggered && tc.notifier.lastReceived == (notifier.Payload{}) {
				t.Error("Expected notifier to be triggered")
			}

			if tc.notifierTriggered && reflect.DeepEqual(tc.notifier.lastReceived, tc.arg) {
				t.Error("Notifier was triggered with wrong payload")
			}
		})
	}
}
