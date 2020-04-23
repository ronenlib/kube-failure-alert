package handler

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ronenlib/kube-event-alert/pkg/notifier"
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

func getExpectedPayload() notifier.Payload {
	return notifier.Payload{
		Kind:      "kind",
		Namespace: "namespace",
		Name:      "name",
		Error:     "reason message",
	}
}

func TestHandle(t *testing.T) {
	cases := []struct {
		notifier           *mockNotifier
		arg                interface{}
		expectNotifierCall bool
		expectPayload      notifier.Payload
		expectError        bool
		name               string
	}{
		{
			notifier:           newMockNotifier(false),
			arg:                newEvent(corev1.EventTypeNormal),
			expectNotifierCall: false,
			expectError:        false,
			name:               "NormalEvent",
		},
		{
			notifier:           newMockNotifier(false),
			arg:                newEvent(corev1.EventTypeWarning),
			expectNotifierCall: true,
			expectPayload:      getExpectedPayload(),
			expectError:        false,
			name:               "WarningEvent",
		},
		{
			notifier:           newMockNotifier(false),
			arg:                struct{}{},
			expectNotifierCall: false,
			expectError:        true,
			name:               "BadType",
		},
		{
			notifier:           newMockNotifier(true),
			arg:                newEvent(corev1.EventTypeWarning),
			expectNotifierCall: true,
			expectPayload:      getExpectedPayload(),
			expectError:        true,
			name:               "ErrorNotifier",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewEventHandler(tc.notifier)
			err := h.Handle(tc.arg)
			receivedErr := err != nil

			if receivedErr != tc.expectError {
				t.Errorf("Expected error to be %v but received error was %v", tc.expectError, receivedErr)
			}

			notifierCalled := tc.notifier.lastReceived == (notifier.Payload{})

			if tc.expectNotifierCall == notifierCalled {
				t.Errorf("Expected notifier called to be %v but got %v", tc.expectNotifierCall, notifierCalled)
			}

			if tc.expectNotifierCall && !reflect.DeepEqual(tc.notifier.lastReceived, tc.expectPayload) {
				t.Error("Notifier was triggered with wrong payload")
			}
		})
	}
}
