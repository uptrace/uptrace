package bunapp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/segmentio/encoding/json"

	"github.com/prometheus/alertmanager/api/v2/models"
)

type Notifier struct {
	client *http.Client
	urls   []string
}

func NewNotifier(urls []string) *Notifier {
	return &Notifier{
		client: http.DefaultClient,
		urls:   urls,
	}
}

func (n *Notifier) Send(ctx context.Context, alerts models.PostableAlerts) error {
	if len(n.urls) == 0 {
		return nil
	}

	payload, err := json.Marshal(alerts)
	if err != nil {
		return err
	}

	var firstErr error

	for _, url := range n.urls {
		if err := n.post(ctx, url, payload); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (n *Notifier) post(ctx context.Context, url string, b []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Any HTTP status 2xx is OK.
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("bad response status %s: %.100q", resp.Status, body)
	}
	return nil
}
