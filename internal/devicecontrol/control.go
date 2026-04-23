package devicecontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	ActionStatus = "status"
	ActionReboot = "reboot"
)

const (
	VendorGenericHTTP = "generic-http"
	VendorTPLINKHTTP  = "tp-link-http"
)

// Request describes a single control operation.
type Request struct {
	Action      string
	TargetURL   string
	Vendor      string
	Username    string
	Password    string
	InsecureTLS bool
	Timeout     time.Duration
}

// Response describes a control result.
type Response struct {
	Action     string
	TargetURL  string
	Success    bool
	StatusCode int
	Message    string
}

type adapter interface {
	buildEndpoint(baseURL string, action string) (string, error)
}

type genericHTTPAdapter struct{}

func (genericHTTPAdapter) buildEndpoint(baseURL string, action string) (string, error) {
	return strings.TrimRight(baseURL, "/") + "/api/" + action, nil
}

type tplinkHTTPAdapter struct{}

func (tplinkHTTPAdapter) buildEndpoint(baseURL string, action string) (string, error) {
	switch action {
	case ActionStatus:
		return strings.TrimRight(baseURL, "/") + "/api/system/status", nil
	case ActionReboot:
		return strings.TrimRight(baseURL, "/") + "/api/system/reboot", nil
	default:
		return "", fmt.Errorf("unsupported action for tp-link-http: %s", action)
	}
}

func resolveAdapter(vendor string) (adapter, error) {
	switch strings.ToLower(strings.TrimSpace(vendor)) {
	case "", VendorGenericHTTP:
		return genericHTTPAdapter{}, nil
	case VendorTPLINKHTTP:
		return tplinkHTTPAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported vendor adapter: %s", vendor)
	}
}

// Execute runs a control action for known vendor adapters.
func Execute(ctx context.Context, req Request) (Response, error) {
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.TargetURL = strings.TrimSpace(req.TargetURL)
	req.Vendor = strings.TrimSpace(req.Vendor)
	if req.Vendor == "" {
		req.Vendor = VendorGenericHTTP
	}
	if req.Timeout <= 0 {
		req.Timeout = 10 * time.Second
	}
	if req.Action != ActionStatus && req.Action != ActionReboot {
		return Response{}, fmt.Errorf("unsupported action: %s", req.Action)
	}
	if req.TargetURL == "" {
		return Response{}, fmt.Errorf("target URL is required")
	}
	if !strings.HasPrefix(strings.ToLower(req.TargetURL), "http://") && !strings.HasPrefix(strings.ToLower(req.TargetURL), "https://") {
		return Response{}, fmt.Errorf("target URL must start with http:// or https://")
	}

	ad, err := resolveAdapter(req.Vendor)
	if err != nil {
		return Response{}, err
	}
	endpoint, err := ad.buildEndpoint(req.TargetURL, req.Action)
	if err != nil {
		return Response{}, err
	}
	payload := map[string]string{
		"action": req.Action,
		"vendor": req.Vendor,
	}
	bodyBytes, _ := json.Marshal(payload)

	httpClient := &http.Client{Timeout: req.Timeout}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(req.Username) != "" {
		httpReq.SetBasicAuth(req.Username, req.Password)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("execute action: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	out := Response{
		Action:     req.Action,
		TargetURL:  req.TargetURL,
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		StatusCode: resp.StatusCode,
		Message:    msg,
	}
	if !out.Success {
		return out, fmt.Errorf("device action failed: status=%d", resp.StatusCode)
	}
	return out, nil
}
