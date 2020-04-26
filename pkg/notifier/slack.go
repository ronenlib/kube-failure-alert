package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const contentType = "application/json"

// WebhookNotifier sends notifications to webhook
type WebhookNotifier struct {
	webhookURL string
}

// NewWebhookNotifier creates new webhook notifier
func NewWebhookNotifier(webhookURL string) WebhookNotifier {
	return WebhookNotifier{
		webhookURL: webhookURL,
	}
}

// Notify convers payload to a readable message
// and sends it to webhook webhook url
func (sn WebhookNotifier) Notify(payload Payload) error {
	reader, err := sn.toReader(payload)

	if err != nil {
		return err
	}

	resp, err := http.Post(sn.webhookURL, contentType, reader)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to send notification to webhook, status %d", resp.StatusCode)
	}

	return nil
}

func (sn WebhookNotifier) toReader(payload Payload) (io.Reader, error) {
	text := fmt.Sprintf("%s %s/%s - %s", payload.Kind, payload.Namespace, payload.Name, payload.Error)

	var body = map[string]string{
		"text": text,
	}

	rawBody, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(rawBody), nil
}
