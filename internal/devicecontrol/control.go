package devicecontrol

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ActionStatus = "status"
	ActionReboot = "reboot"
)

// ConsentToken — обязательный токен подтверждения необратимого действия.
//
// P0-3: подтверждение проверяется не только в CLI, но и здесь, чтобы ни один
// вызывающий слой (CLI, GUI, REST) не мог выполнить reboot «случайно».
const ConsentToken = "I_UNDERSTAND"

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

	// Consent — токен подтверждения для необратимых действий (reboot).
	// Для ActionReboot обязателен ConsentToken.
	Consent string
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

// validateTargetURL строго проверяет URL устройства управления.
//
// Правила (защита от ввода произвольных/небезопасных целей):
//   - разрешены только схемы http и https;
//   - обязательны host и (для http/https) отсутствие user-info, чтобы учётные
//     данные не попадали в URL/логи (для аутентификации используется SetBasicAuth);
//   - запрещены управляющие символы и пробелы, которые ломают разбор адреса.
func validateTargetURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("target URL is required")
	}
	// Проверка управляющих символов идёт по исходной строке ДО TrimSpace:
	// иначе перевод строки или табуляция на конце адреса будут молча срезаны
	// и останутся незамеченными, хотя такой ввод считается небезопасным.
	if strings.ContainsAny(raw, " \t\r\n") {
		return fmt.Errorf("target URL must not contain whitespace")
	}
	raw = trimmed
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("target URL must start with http:// or https://")
	}
	if strings.TrimSpace(u.Host) == "" {
		return fmt.Errorf("target URL must contain a host")
	}
	if u.User != nil {
		return fmt.Errorf("target URL must not contain credentials; use username/password fields")
	}
	return nil
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
	// P0-3: reboot подтверждается до обращения к устройству. Проверка на уровне
	// сервиса, а не только CLI/GUI, исключает обход подтверждения новым
	// вызывающим слоем.
	if req.Action == ActionReboot && strings.TrimSpace(req.Consent) != ConsentToken {
		return Response{Action: req.Action, TargetURL: req.TargetURL},
			fmt.Errorf("reboot requires explicit confirmation: consent must be %s", ConsentToken)
	}
	if err := validateTargetURL(req.TargetURL); err != nil {
		return Response{}, err
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
	// InsecureTLS — осознанный обход проверки сертификата для устройств с
	// самоподписанными сертификатами (домашние роутеры, старые коммутаторы).
	// По умолчанию выключен: GUI не включает его, поэтому обычные запросы
	// проходят полную проверку цепочки. Минимум TLS 1.2.
	if req.InsecureTLS {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // осознанно: самоподписанные сертификаты устройств; доступно только при явном InsecureTLS, GUI по умолчанию выключен
				MinVersion:         tls.VersionTLS12,
			},
		}
	}
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
