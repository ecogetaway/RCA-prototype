package keep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s.io/klog"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      *time.Time        `json:"endsAt,omitempty"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendAlert(alert Alert) error {
	if c.baseURL == "" {
		// Mock: just log the alert
		klog.Infof("Mock Keep Alert: ID=%s, Name=%s, Status=%s", alert.ID, alert.Name, alert.Status)
		return nil
	}

	url := fmt.Sprintf("%s/alerts", c.baseURL)
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send alert: %s", resp.Status)
	}

	klog.Infof("Alert sent to Keep: %s", alert.ID)
	return nil
}