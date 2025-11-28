package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/dashboard/models"
)

type Client struct {
	baseURL string
	http    *http.Client
}

// New creates a new API client
// If API_URL is not set, it uses relative paths for same-origin calls
func New() *Client {
	base := os.Getenv("API_URL")
	// Leave baseURL empty for same-origin calls in production
	return &Client{
		baseURL: base,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// makeURL constructs the full URL for API calls
// If baseURL is empty, returns the path as-is for same-origin calls
func (c *Client) makeURL(path string) string {
	if c.baseURL == "" {
		return path
	}
	return strings.TrimRight(c.baseURL, "/") + path
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.makeURL("/alerts/"+url.PathEscape(alertID)+"/acknowledge"), nil)
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
	payload := models.AnalyticsGenerateRequest{FacilityID: facilityID, Date: date}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.makeURL("/analytics/generate"), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("generate analytics failed: %s", resp.Status)
	}
	var out models.AnalyticsGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any, params url.Values) error {
	u := c.makeURL(path)
	if params != nil {
		if strings := params.Encode(); strings != "" {
			u += "?" + strings
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
