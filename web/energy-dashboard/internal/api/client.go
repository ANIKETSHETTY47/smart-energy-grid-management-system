package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"energy-dashboard-go/internal/models"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New() *Client {
	base := os.Getenv("API_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Client{
		baseURL: base,
		http: &http.Client{ Timeout: 10 * time.Second },
	}
}

func (c *Client) Health(ctx context.Context) (*models.Health, error) {
	var out models.Health
	if err := c.getJSON(ctx, "/health", &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Facilities(ctx context.Context) ([]models.Facility, error) {
	var out []models.Facility
	if err := c.getJSON(ctx, "/facilities", &out, nil); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) RecentReadings(ctx context.Context, facilityID string, hours int) (*models.RecentReadingsResponse, error) {
	params := url.Values{}
	params.Set("facility_id", facilityID)
	params.Set("hours", fmt.Sprintf("%d", hours))
	var out models.RecentReadingsResponse
	if err := c.getJSON(ctx, "/readings/recent", &out, params); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Alerts(ctx context.Context, facilityID, severity string) (*models.AlertsResponse, error) {
	params := url.Values{}
	params.Set("facility_id", facilityID)
	if severity != "" {
		params.Set("severity", severity)
	}
	var out models.AlertsResponse
	if err := c.getJSON(ctx, "/alerts", &out, params); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AcknowledgeAlert(ctx context.Context, alertID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/alerts/"+url.PathEscape(alertID)+"/acknowledge", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("acknowledge failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) GenerateAnalytics(ctx context.Context, facilityID, date string) (*models.AnalyticsGenerateResponse, error) {
	payload := models.AnalyticsGenerateRequest{ FacilityID: facilityID, Date: date }
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analytics/generate", bytes.NewReader(b))
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("generate analytics failed: %s", resp.Status)
	}
	var out models.AnalyticsGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any, params url.Values) error {
	u := c.baseURL + path
	if params != nil {
		if strings := params.Encode(); strings != "" {
			u += "?" + strings
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil { return err }
	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
