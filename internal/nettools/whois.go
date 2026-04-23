package nettools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

type rdapVCardItem struct {
	Name   string
	Type   any
	Format string
	Value  any
}

type rdapEntity struct {
	Handle []string `json:"roles"`
	VCard  []any    `json:"vcardArray"`
}

type rdapResponse struct {
	ObjectClassName string       `json:"objectClassName"`
	Handle          string       `json:"handle"`
	LDHName         string       `json:"ldhName"`
	Name            string       `json:"name"`
	StartAddress    string       `json:"startAddress"`
	EndAddress      string       `json:"endAddress"`
	Country         string       `json:"country"`
	Status          []string     `json:"status"`
	Events          []rdapEvent  `json:"events"`
	Entities        []rdapEntity `json:"entities"`
}

// RunWhois вызывает внешнюю утилиту whois, если она есть в PATH.
func RunWhois(ctx context.Context, query string, timeout time.Duration) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("пустой запрос")
	}
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	s, err := runCmd(ctx, []string{"whois", query}, timeout)
	if err != nil {
		var te *ToolError
		if errors.As(err, &te) && te.Code == ToolErrorNotInstalled {
			return runWhoisRDAP(ctx, query, timeout)
		}
		return "", err
	}
	if runtime.GOOS == "windows" {
		return s + "\n\n(Windows: для полного whois может потребоваться установка клиента, например Sysinternals или пакет whois)", nil
	}
	return s, nil
}

func runWhoisRDAP(ctx context.Context, query string, timeout time.Duration) (string, error) {
	rdapURL := buildRDAPURL(query)
	if strings.TrimSpace(rdapURL) == "" {
		return "", fmt.Errorf("не удалось построить RDAP URL для запроса %q", query)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, rdapURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", newToolError("whois", ToolErrorNetwork, "ошибка обращения к RDAP сервису", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", newToolError("whois", ToolErrorNetwork, fmt.Sprintf("RDAP вернул HTTP %d", resp.StatusCode), nil)
	}
	if summary, ok := formatRDAPSummary(body, query); ok {
		return "whois утилита не найдена; используется RDAP fallback.\n\n" + summary, nil
	}
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		raw = "RDAP вернул пустой ответ"
	}
	return "whois утилита не найдена; используется RDAP fallback.\n\n" + raw, nil
}

func buildRDAPURL(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return ""
	}
	base := resolveRDAPBaseURL()
	if ip := net.ParseIP(q); ip != nil {
		return base + "/ip/" + url.PathEscape(q)
	}
	return base + "/domain/" + url.PathEscape(strings.ToLower(q))
}

func resolveRDAPBaseURL() string {
	const defaultBase = "https://rdap.org"
	raw := strings.TrimSpace(os.Getenv("NETWORK_SCANNER_RDAP_BASE_URL"))
	if raw == "" {
		return defaultBase
	}
	raw = strings.TrimRight(raw, "/")
	if _, err := url.ParseRequestURI(raw); err != nil {
		return defaultBase
	}
	return raw
}

func compactRDAP(body []byte) (string, bool) {
	var any map[string]any
	if err := json.Unmarshal(body, &any); err != nil {
		return "", false
	}
	out, err := json.MarshalIndent(any, "", "  ")
	if err != nil {
		return "", false
	}
	return string(out), true
}

func formatRDAPSummary(body []byte, query string) (string, bool) {
	var rr rdapResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return "", false
	}
	var sb strings.Builder
	sb.WriteString("RDAP summary\n")
	sb.WriteString(fmt.Sprintf("- Query: %s\n", strings.TrimSpace(query)))
	if v := firstNonEmpty(rr.LDHName, rr.Name, rr.Handle); v != "" {
		sb.WriteString(fmt.Sprintf("- Name: %s\n", v))
	}
	if strings.TrimSpace(rr.ObjectClassName) != "" {
		sb.WriteString(fmt.Sprintf("- Type: %s\n", strings.TrimSpace(rr.ObjectClassName)))
	}
	if strings.TrimSpace(rr.Country) != "" {
		sb.WriteString(fmt.Sprintf("- Country: %s\n", strings.TrimSpace(rr.Country)))
	}
	if strings.TrimSpace(rr.StartAddress) != "" || strings.TrimSpace(rr.EndAddress) != "" {
		sb.WriteString(fmt.Sprintf("- Range: %s - %s\n", strings.TrimSpace(rr.StartAddress), strings.TrimSpace(rr.EndAddress)))
	}
	if len(rr.Status) > 0 {
		sb.WriteString(fmt.Sprintf("- Status: %s\n", strings.Join(cleanSlice(rr.Status), ", ")))
	}
	if registrar := extractRegistrar(rr.Entities); registrar != "" {
		sb.WriteString(fmt.Sprintf("- Registrar/Org: %s\n", registrar))
	}
	if created, updated := extractDates(rr.Events); created != "" || updated != "" {
		if created != "" {
			sb.WriteString(fmt.Sprintf("- Created: %s\n", created))
		}
		if updated != "" {
			sb.WriteString(fmt.Sprintf("- Updated: %s\n", updated))
		}
	}
	if strings.TrimSpace(sb.String()) == "RDAP summary" {
		return "", false
	}
	return strings.TrimSpace(sb.String()), true
}

func extractDates(events []rdapEvent) (created string, updated string) {
	for _, e := range events {
		action := strings.ToLower(strings.TrimSpace(e.EventAction))
		date := strings.TrimSpace(e.EventDate)
		if date == "" {
			continue
		}
		switch action {
		case "registration", "created":
			if created == "" {
				created = date
			}
		case "last changed", "last update of rdap database", "updated":
			if updated == "" {
				updated = date
			}
		}
	}
	return created, updated
}

func extractRegistrar(entities []rdapEntity) string {
	for _, e := range entities {
		if hasRole(e.Handle, "registrar") || hasRole(e.Handle, "registrant") {
			if name := extractEntityName(e.VCard); name != "" {
				return name
			}
		}
	}
	for _, e := range entities {
		if name := extractEntityName(e.VCard); name != "" {
			return name
		}
	}
	return ""
}

func hasRole(roles []string, role string) bool {
	want := strings.ToLower(strings.TrimSpace(role))
	for _, r := range roles {
		if strings.ToLower(strings.TrimSpace(r)) == want {
			return true
		}
	}
	return false
}

func extractEntityName(vcard []any) string {
	if len(vcard) < 2 {
		return ""
	}
	items, ok := vcard[1].([]any)
	if !ok {
		return ""
	}
	for _, raw := range items {
		arr, ok := raw.([]any)
		if !ok || len(arr) < 4 {
			continue
		}
		fieldName, _ := arr[0].(string)
		if strings.ToLower(strings.TrimSpace(fieldName)) != "fn" {
			continue
		}
		if s, ok := arr[3].(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func cleanSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
