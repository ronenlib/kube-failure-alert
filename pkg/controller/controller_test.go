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

type MockHandler struct {
	withError bool
}

func newMockHandler(withError bool) MockHandler {
	return MockHandler{
		withError: withError,
	}
}

func (s *MockHandler) Handle(obj interface{}) error {
	if s.withError {
		return errors.New("failed to handle")
	}

	return nil
}

func newPod(name, namespace string, phase corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: corev1.PodStatus{
			Phase: phase,
		},
	}
}

type fixture struct {
	controller      *Controller
	handler         *MockHandler
	informerFactory informers.SharedInformerFactory
	stopCh          chan struct{}
}

func newFixture(handler MockHandler, pods ...*corev1.Pod) *fixture {
	client := fake.NewSimpleClientset()

	factory := informers.NewSharedInformerFactory(client, 0)

	f := &fixture{}
	f.stopCh = make(chan struct{})
	f.handler = &handler
	f.informerFactory = factory

	informer := f.informerFactory.Core().V1().Pods()
	f.controller = newController("pod-test", client, informer, &handler)

	f.informerFactory.Start(f.stopCh)
	f.informerFactory.WaitForCacheSync(f.stopCh)

	for _, p := range pods {
		_ = f.controller.informer.GetIndexer().Add(p)
	}

	return f
}

func (f *fixture) stop() {
	close(f.stopCh)
}

func TestControllerProcessItem(t *testing.T) {
	cases := []struct {
		handler        MockHandler
		pod            *corev1.Pod
		key            interface{}
		expectedResult bool
		expectError    bool
		name           string
	}{
		{
			handler:        newMockHandler(false),
			pod:            newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:            fmt.Sprintf("%s/pod", metav1.NamespaceDefault),
			expectedResult: true,
			expectError:    false,
			name:           "ValidKey",
		},
		{
			handler:        newMockHandler(false),
			pod:            newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:            -1,
			expectedResult: true,
			expectError:    true,
			name:           "InvalidKey",
		},
		{
			handler:        newMockHandler(true),
			pod:            newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:            fmt.Sprintf("%s/pod", metav1.NamespaceDefault),
			expectedResult: true,
			expectError:    true,
			name:           "HandlerError",
		},
	}

	for _, tc := range cases {
		f := newFixture(tc.handler, tc.pod)
		defer f.stop()

		t.Run(tc.name, func(t *testing.T) {
			f.controller.workqueue.Add(tc.key)
			result, err := f.controller.processNextWorkItem()

			if result != tc.expectedResult {
				t.Errorf("Expected return value to be %v, got %v", tc.expectedResult, result)
			}

			if err != nil && !tc.expectError {
				t.Error("Expected error")
			}
		})
	}
}

func TestControllerHandelKey(t *testing.T) {
	cases := []struct {
		handler     MockHandler
		pod         *corev1.Pod
		key         string
		expectError bool
		name        string
	}{
		{
			handler:     newMockHandler(false),
			pod:         newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:         fmt.Sprintf("%s/pod", metav1.NamespaceDefault),
			expectError: false,
			name:        "ValidKey",
		},
		{
			handler:     newMockHandler(false),
			pod:         newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:         fmt.Sprintf("%s/unknown", metav1.NamespaceDefault),
			expectError: true,
			name:        "UnknownKey",
		},
		{
			handler:     newMockHandler(false),
			pod:         newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:         "pod",
			expectError: true,
			name:        "InvalidKey",
		},
		{
			handler:     newMockHandler(true),
			pod:         newPod("pod", metav1.NamespaceDefault, corev1.PodRunning),
			key:         "pod",
			expectError: true,
			name:        "HandlerError",
		},
	}

	for _, tc := range cases {
		f := newFixture(tc.handler, tc.pod)
		defer f.stop()

		t.Run(tc.name, func(t *testing.T) {
			err := f.controller.handleKey(tc.key)

			if err != nil && !tc.expectError {
				t.Error("Expected error")
			}
		})
	}
}
