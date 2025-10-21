package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPError struct {
	Status int
	Method string
	URL    string
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s %s failed: %d - %s", e.Method, e.URL, e.Status, e.Body)
}

type Client struct {
	base   string
	apiKey string
	http   *http.Client
}

func NewClient(base, apiKey string) *Client {
	return &Client{
		base:   strings.TrimRight(base, "/"),
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	u := c.base + path

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}

	// Header name per your middleware note:
	req.Header.Set("x-rd-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	}

	b, _ := io.ReadAll(resp.Body)
	return &HTTPError{
		Status: resp.StatusCode,
		Method: method,
		URL:    u,
		Body:   string(b),
	}
}

type JobStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ServerOS struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ServerPlan struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	RAMGB   int    `json:"ram"`
	Storage int    `json:"storage"`
	CPUName string `json:"cpuName"`
	Cores   int    `json:"cores"`
}

type ServerLocation struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
	Country string `json:"country"`
}

type CreateServerRequest struct {
	PlanID     int     `json:"planId"`
	LocationID int     `json:"locationId"`
	OSID       int     `json:"osId"`
	Raid       *int    `json:"raid,omitempty"`
	Hostname   *string `json:"hostname,omitempty"`
}

type Server struct {
	ID           string         `json:"id,omitempty"`
	Plan         ServerPlan     `json:"plan"`
	Location     ServerLocation `json:"location"`
	ServerOS     *ServerOS      `json:"serverOS,omitempty"`
	Raid         *int           `json:"raid,omitempty"`
	Hostname     *string        `json:"hostname,omitempty"`
	IPAddress    string         `json:"ipAddress,omitempty"`
	PowerStatus  *string        `json:"devicePowerStatus,omitempty"`
	MonthlyPrice *string        `json:"monthlyPrice,omitempty"`
}

type ServerListItem struct {
	ID          string  `json:"id,omitempty"`
	Hostname    *string `json:"hostname,omitempty"`
	IPAddress   string  `json:"ipAddress,omitempty"`
	PowerStatus *string `json:"powerStatus,omitempty"`
}

type CPU struct {
	Name  string  `json:"name"`
	Cores int     `json:"cores"`
	Speed float64 `json:"speedGhz"`
}

type PlanLocation struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Keyword      string `json:"keyword"`
	MonthlyPrice int    `json:"monthlyPrice"`
}

type Plan struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	CPU       CPU            `json:"cpu"`
	Locations []PlanLocation `json:"locations"`
	RAMGB     int            `json:"ram"`
	Storage   int            `json:"storageGb"`
}

// //////
// Responses from api
// //////
type EnvelopeServer struct {
	Success    bool   `json:"success"`
	Data       Server `json:"data"`
	Message    string `json:"message"`
	TotalCount int    `json:"totalCount,omitempty"`
}

type EnvelopeServerListItem struct {
	Success    bool           `json:"success"`
	Data       ServerListItem `json:"data"`
	Message    string         `json:"message"`
	TotalCount int            `json:"totalCount,omitempty"`
}

type EnvelopePlans struct {
	Success bool   `json:"success"`
	Data    []Plan `json:"data"`
	Message string `json:"message"`
}

type EnvelopeRaidCheck struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type EnvelopeOS struct {
	Success bool       `json:"success"`
	Data    []ServerOS `json:"data"`
	Message string     `json:"message"`
}

func (c *Client) CreateServer(ctx context.Context, reqBody *CreateServerRequest) (*ServerListItem, error) {
	var env EnvelopeServerListItem
	if err := c.do(ctx, http.MethodPost, "/v1/ordering/allocate", reqBody, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("%s", env.Message)
	}
	out := env.Data
	return &out, nil
}

func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	var env EnvelopeServer
	if err := c.do(ctx, http.MethodGet, "/v1/servers/"+url.PathEscape(id), nil, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("%s", env.Message)
	}
	out := env.Data
	return &out, nil
}

func (c *Client) DeleteServer(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/servers/"+url.PathEscape(id)+"/destroy", nil, nil)
}

func (c *Client) ListPlans(ctx context.Context, location string) ([]Plan, error) {
	var env EnvelopePlans
	path := "/v1/ordering/plans?showAll=true"
	if location != "" {
		path += "&location=" + url.QueryEscape(location)
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("%s", env.Message)
	}
	return env.Data, nil
}

func (c *Client) CheckRaid(ctx context.Context, raid int, planID int) (bool, error) {
	var env EnvelopeRaidCheck
	path := fmt.Sprintf("/v1/ordering/plans/%d/raid/%d/check", planID, raid)
	if err := c.do(ctx, http.MethodGet, path, nil, &env); err != nil {
		return false, err
	}
	if !env.Success {
		return false, fmt.Errorf("%s", env.Message)
	}
	return true, nil
}

func (c *Client) ListOperatingSystems(ctx context.Context) ([]ServerOS, error) {
	var env EnvelopeOS
	if err := c.do(ctx, http.MethodGet, "/v1/ordering/os", nil, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("%s", env.Message)
	}
	return env.Data, nil
}
