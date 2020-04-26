package notifier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type webhookRequest struct {
	Text string `json:"text"`
}

var body webhookRequest

func setWebhookBody(req *http.Request) {
	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		panic("failed to parse json")
	}
}

func fakeOkHandler(w http.ResponseWriter, req *http.Request) {
	setWebhookBody(req)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func fakeBadMessageHandler(w http.ResponseWriter, req *http.Request) {
	setWebhookBody(req)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "error")
}

func TestNotify(t *testing.T) {
	cases := []struct {
		handler     http.Handler
		payload     Payload
		expectError bool
		name        string
	}{
		{
			handler: http.HandlerFunc(fakeOkHandler),
			payload: Payload{
				Kind:      "kind",
				Namespace: "namespace",
				Name:      "name",
				Error:     "error",
			},
			expectError: false,
			name:        "NotifySuccessful",
		},
		{
			handler:     http.HandlerFunc(fakeBadMessageHandler),
			payload:     Payload{},
			expectError: true,
			name:        "NotifyFailure",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			sn := NewWebhookNotifier(server.URL)

			body = webhookRequest{} // reset variable
			err := sn.Notify(tc.payload)
			receivedErr := err != nil

			text := fmt.Sprintf("%s %s/%s - %s", tc.payload.Kind, tc.payload.Namespace, tc.payload.Name, tc.payload.Error)

			if text != body.Text {
				t.Errorf("Notifier sent the wrong message")
			}

			if receivedErr != tc.expectError {
				t.Errorf("Expected error to be %v but received error was %v", tc.expectError, receivedErr)
			}
		})
	}
}
