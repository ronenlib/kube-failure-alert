package controller

import (
	"errors"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

type mockHandler struct {
	withError bool
}

func newMockHandler(withError bool) mockHandler {
	return mockHandler{
		withError: withError,
	}
}

func (s *mockHandler) Handle(obj interface{}) error {
	if s.withError {
		return errors.New("failed to handle")
	}

	return nil
}

func newEvent(name, namespace string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

type fixture struct {
	controller      *Controller
	handler         *mockHandler
	informerFactory informers.SharedInformerFactory
	stopCh          chan struct{}
}

func newFixture(handler mockHandler, events ...*corev1.Event) *fixture {
	client := fake.NewSimpleClientset()

	factory := informers.NewSharedInformerFactory(client, 0)

	f := &fixture{}
	f.stopCh = make(chan struct{})
	f.handler = &handler
	f.informerFactory = factory

	informer := f.informerFactory.Core().V1().Events()
	f.controller = newController("event-test", client, informer, &handler)

	f.informerFactory.Start(f.stopCh)
	f.informerFactory.WaitForCacheSync(f.stopCh)

	for _, e := range events {
		_ = f.controller.informer.GetIndexer().Add(e)
	}

	return f
}

func (f *fixture) stop() {
	close(f.stopCh)
}

func TestControllerProcessItem(t *testing.T) {
	cases := []struct {
		handler        mockHandler
		event          *corev1.Event
		key            interface{}
		expectedResult bool
		expectError    bool
		name           string
	}{
		{
			handler:        newMockHandler(false),
			event:          newEvent("event", metav1.NamespaceDefault),
			key:            fmt.Sprintf("%s/event", metav1.NamespaceDefault),
			expectedResult: true,
			expectError:    false,
			name:           "ValidKey",
		},
		{
			handler:        newMockHandler(false),
			event:          newEvent("event", metav1.NamespaceDefault),
			key:            -1,
			expectedResult: true,
			expectError:    true,
			name:           "InvalidKey",
		},
		{
			handler:        newMockHandler(true),
			event:          newEvent("event", metav1.NamespaceDefault),
			key:            fmt.Sprintf("%s/event", metav1.NamespaceDefault),
			expectedResult: true,
			expectError:    true,
			name:           "HandlerError",
		},
	}

	for _, tc := range cases {
		f := newFixture(tc.handler, tc.event)
		defer f.stop()

		t.Run(tc.name, func(t *testing.T) {
			f.controller.workqueue.Add(tc.key)
			result, err := f.controller.processNextWorkItem()
			receivedErr := err != nil

			if result != tc.expectedResult {
				t.Errorf("Expected return value to be %v, got %v", tc.expectedResult, result)
			}

			if receivedErr != tc.expectError {
				t.Errorf("Expected error to be %v but received error was %v", tc.expectError, receivedErr)
			}
		})
	}
}

func TestControllerHandelKey(t *testing.T) {
	cases := []struct {
		handler     mockHandler
		event       *corev1.Event
		key         string
		expectError bool
		name        string
	}{
		{
			handler:     newMockHandler(false),
			event:       newEvent("event", metav1.NamespaceDefault),
			key:         fmt.Sprintf("%s/event", metav1.NamespaceDefault),
			expectError: false,
			name:        "ValidKey",
		},
		{
			handler:     newMockHandler(false),
			event:       newEvent("event", metav1.NamespaceDefault),
			key:         fmt.Sprintf("%s/unknown", metav1.NamespaceDefault),
			expectError: true,
			name:        "UnknownKey",
		},
		{
			handler:     newMockHandler(false),
			event:       newEvent("event", metav1.NamespaceDefault),
			key:         "event",
			expectError: true,
			name:        "InvalidKey",
		},
		{
			handler:     newMockHandler(true),
			event:       newEvent("event", metav1.NamespaceDefault),
			key:         "event",
			expectError: true,
			name:        "HandlerError",
		},
	}

	for _, tc := range cases {
		f := newFixture(tc.handler, tc.event)
		defer f.stop()

		t.Run(tc.name, func(t *testing.T) {
			err := f.controller.handleKey(tc.key)
			receivedErr := err != nil

			if receivedErr != tc.expectError {
				t.Errorf("Expected error to be %v but received error was %v", tc.expectError, receivedErr)
			}
		})
	}
}
