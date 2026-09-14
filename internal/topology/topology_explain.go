package topology

import (
	"fmt"
	"sort"
	"strings"
)

// DedupReport возвращает текстовый отчёт о связях топологии: сколько устройств и
// связей, как связи распределены по источникам и достоверности.
//
// Отчёт диагностический — нужен, чтобы глазами увидеть, какие источники дали
// ссылки и не затирает ли более слабый источник более сильный. nil-топология
// не паникует: отчёт вызывается из обработчиков ошибок и логов.
func (t *Topology) DedupReport() string {
	if t == nil {
		return "Topology Deduplication Report: <nil topology>\n"
	}

	bySource := make(map[LinkSourceType]int, 3)
	byConfidence := make(map[LinkConfidence]int, 3)
	for _, l := range t.Links {
		bySource[l.SourceType]++
		byConfidence[l.Confidence]++
	}

	var b strings.Builder
	b.WriteString("Topology Deduplication Report:\n")
	fmt.Fprintf(&b, "  Total devices: %d\n", len(t.Devices))
	fmt.Fprintf(&b, "  Total links: %d\n", len(t.Links))

	b.WriteString("  Links by source type:\n")
	for _, source := range sortedSourceTypes(bySource) {
		fmt.Fprintf(&b, "    %s: %d\n", source, bySource[source])
	}

	b.WriteString("  Links by confidence:\n")
	for _, conf := range sortedConfidences(byConfidence) {
		fmt.Fprintf(&b, "    %s: %d\n", conf, byConfidence[conf])
	}

	return b.String()
}

// sortedSourceTypes перечисляет источники связей по убыванию приоритета,
// неизвестные источники — по имени. Порядок детерминирован, чтобы отчёт можно
// было сравнивать между запусками.
func sortedSourceTypes(counts map[LinkSourceType]int) []LinkSourceType {
	known := []LinkSourceType{LinkSourceLLDP, LinkSourceFDB, LinkSourceInferred}
	out := make([]LinkSourceType, 0, len(counts))
	for _, source := range known {
		if counts[source] > 0 {
			out = append(out, source)
		}
	}
	var extra []string
	for source := range counts {
		if source != LinkSourceLLDP && source != LinkSourceFDB && source != LinkSourceInferred {
			extra = append(extra, string(source))
		}
	}
	sort.Strings(extra)
	for _, source := range extra {
		out = append(out, LinkSourceType(source))
	}
	return out
}

// sortedConfidences перечисляет уровни достоверности от высокой к низкой.
func sortedConfidences(counts map[LinkConfidence]int) []LinkConfidence {
	known := []LinkConfidence{LinkConfidenceHigh, LinkConfidenceMedium, LinkConfidenceLow}
	out := make([]LinkConfidence, 0, len(counts))
	for _, conf := range known {
		if counts[conf] > 0 {
			out = append(out, conf)
		}
	}
	var extra []string
	for conf := range counts {
		if conf != LinkConfidenceHigh && conf != LinkConfidenceMedium && conf != LinkConfidenceLow {
			extra = append(extra, string(conf))
		}
	}
	sort.Strings(extra)
	for _, conf := range extra {
		out = append(out, LinkConfidence(conf))
	}
	return out
}

// ExplainLink человекочитательно объясняет, откуда взялась связь с указанным
// номером. Возвращает false, если индекс вне диапазона — вызывающий код GUI
// показывает сообщение об ошибке вместо паники.
func (t *Topology) ExplainLink(index int) (string, bool) {
	if t == nil || index < 0 || index >= len(t.Links) {
		return "", false
	}

	link := t.Links[index]
	evidence := strings.TrimSpace(link.Evidence)
	sourceType := strings.ToUpper(string(link.SourceType))
	if link.SourceType == "" {
		sourceType = string(LinkSourceInferred)
	}

	var detail string
	switch link.SourceType {
	case LinkSourceLLDP:
		detail = explainLLDPSource(link.Source, link.Target, portLabel(link.SourcePort), evidence)
	case LinkSourceFDB:
		detail = explainFDBSource(link.Source, link.Target, linkMAC(link), evidence)
	default:
		detail = explainInferredSource(link.Source, link.Target, evidence)
	}

	head := fmt.Sprintf("Связь #%d: %s -> %s [%s, достоверность: %s]",
		index+1, deviceDisplayName(link.Source), deviceDisplayName(link.Target), sourceType, link.Confidence)
	if detail == "" {
		return head, true
	}
	return head + "\n" + detail, true
}

// explainLLDPSource объясняет связь из LLDP-объявления соседа.
func explainLLDPSource(src, dst *Device, portName, evidence string) string {
	if portName == "" {
		portName = "неизвестный порт"
	}
	if evidence == "" {
		evidence = "нет данных"
	}
	return fmt.Sprintf("LLDP: %s объявил соседа %s на порту %s (evidence: %s)",
		deviceDisplayName(src), deviceDisplayName(dst), portName, evidence)
}

// explainFDBSource объясняет связь по таблице MAC коммутатора.
func explainFDBSource(src, dst *Device, mac, evidence string) string {
	if mac == "" {
		mac = "неизвестный MAC"
	}
	if evidence == "" {
		evidence = "нет данных"
	}
	return fmt.Sprintf("FDB: MAC %s (%s) замечен на порту коммутатора %s (evidence: %s)",
		mac, deviceDisplayName(dst), deviceDisplayName(src), evidence)
}

// explainInferredSource объясняет эвристическую связь.
func explainInferredSource(src, dst *Device, reason string) string {
	if reason == "" {
		reason = "причина не указана"
	}
	return fmt.Sprintf("Эвристика: %s и %s связаны по признаку %s",
		deviceDisplayName(src), deviceDisplayName(dst), reason)
}

// linkMAC возвращает MAC-адрес удалённого конца связи — тот, по которому
// строится соответствие FDB-записи и порта.
func linkMAC(link Link) string {
	if link.Target != nil && strings.TrimSpace(link.Target.MAC) != "" {
		return strings.TrimSpace(link.Target.MAC)
	}
	if link.Source != nil {
		return strings.TrimSpace(link.Source.MAC)
	}
	return ""
}
