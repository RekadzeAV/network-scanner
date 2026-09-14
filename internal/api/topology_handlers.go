package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"github.com/gorilla/mux"
)

// topologyRequest — единая схема запроса всех топологических хендлеров.
type topologyRequest struct {
	SnapshotID    string `json:"snapshot_id"`
	SNMPEnabled   bool   `json:"snmp_enabled"`
	SNMPCommunity string `json:"snmp_community"`
	SNMPTimeout   int    `json:"snmp_timeout"`
}

// applyDefaults — значения SNMP-параметров по умолчанию.
func (r *topologyRequest) applyDefaults() {
	if r.SNMPTimeout <= 0 {
		r.SNMPTimeout = 2
	}
	if r.SNMPCommunity == "" {
		r.SNMPCommunity = "public"
	}
}

// errNoSnapshots — снапшоты в inventory отсутствуют.
var errNoSnapshots = errors.New("no snapshots found")

// httpFailure — HTTP-ошибка с кодом статуса и сообщением.
type httpFailure struct {
	code int
	msg  string
	err  error
}

// loadHostsForSnapshot — загрузка хостов снапшота: по ID или последнего.
// Общий для всех топологических хендлеров путь (дедупликация).
func (h *Handler) loadHostsForSnapshot(req topologyRequest) ([]scanner.Result, *httpFailure) {
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		return nil, &httpFailure{http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err), err}
	}
	defer store.Close()

	if req.SnapshotID != "" {
		snap, err := store.LoadSnapshot(req.SnapshotID)
		if err != nil {
			return nil, &httpFailure{http.StatusNotFound, fmt.Sprintf("snapshot not found: %v", err), err}
		}
		return snap.Hosts, nil
	}

	snapshots, err := store.ListSnapshots(1)
	if err != nil {
		return nil, &httpFailure{http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err), err}
	}
	if len(snapshots) == 0 {
		return nil, &httpFailure{http.StatusNotFound, "no snapshots found", errNoSnapshots}
	}
	lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
	if err != nil {
		return nil, &httpFailure{http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err), err}
	}
	return lastSnap.Hosts, nil
}

// buildTopologyForRequest — общий пайплайн: снапшот → SNMP (best-effort) → топология.
func (h *Handler) buildTopologyForRequest(req topologyRequest) (*topology.Topology, *httpFailure) {
	hosts, failure := h.loadHostsForSnapshot(req)
	if failure != nil {
		return nil, failure
	}

	var snmpData map[string]*topology.Device
	if req.SNMPEnabled {
		data, _, err := snmpcollector.CollectWithReport(hosts, []string{req.SNMPCommunity}, req.SNMPTimeout)
		if err != nil {
			// Ошибка SNMP не влияет на ответ: топология строится без доп. данных.
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		}
		snmpData = data
	}

	topo, err := topology.BuildTopology(hosts, snmpData)
	if err != nil {
		return nil, &httpFailure{http.StatusInternalServerError, fmt.Sprintf("build topology: %v", err), err}
	}
	return topo, nil
}

// topologyBuildHandler обрабатывает POST /api/v1/topology/build
func (h *Handler) topologyBuildHandler(w http.ResponseWriter, r *http.Request) {
	var req topologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.applyDefaults()

	topo, failure := h.buildTopologyForRequest(req)
	if failure != nil {
		// Историческое поведение build: пустой inventory — 200 с сообщением.
		if errors.Is(failure.err, errNoSnapshots) {
			h.writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "no snapshots found",
			})
			return
		}
		h.writeError(w, failure.code, failure.msg)
		return
	}

	devices := make([]map[string]interface{}, 0, len(topo.Devices))
	for _, d := range topo.Devices {
		devices = append(devices, map[string]interface{}{
			"ip":             d.IP,
			"mac":            d.MAC,
			"hostname":       d.Hostname,
			"type":           string(d.Type),
			"snmp_enabled":   d.SNMPEnabled,
			"ports":          len(d.Ports),
			"lldp_neighbors": len(d.LldpNeighbors),
		})
	}

	links := make([]map[string]interface{}, 0, len(topo.Links))
	for _, l := range topo.Links {
		links = append(links, map[string]interface{}{
			"source":      deviceDisplayName(l.Source),
			"source_port": portLabel(l.SourcePort),
			"target":      deviceDisplayName(l.Target),
			"target_port": portLabel(l.TargetPort),
			"source_type": string(l.SourceType),
			"confidence":  string(l.Confidence),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"devices":      devices,
		"device_count": len(devices),
		"links":        links,
		"link_count":   len(links),
	})
}

// topologyExportHandler обрабатывает POST /api/v1/topology/export/{format}
func (h *Handler) topologyExportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]

	if format != "json" && format != "dot" && format != "graphml" {
		h.writeError(w, http.StatusBadRequest, "format must be json, dot, or graphml")
		return
	}

	var req topologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.applyDefaults()

	topo, failure := h.buildTopologyForRequest(req)
	if failure != nil {
		h.writeError(w, failure.code, failure.msg)
		return
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(topo, "", "  ")
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("marshal json: %v", err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)

	case "dot":
		w.Header().Set("Content-Type", "text/plain")
		topo.ToDOT(w)

	case "graphml":
		data, err := topo.SaveGraphMLToBytes()
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("marshal graphml: %v", err))
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	}
}

// topologyDOTHandler обрабатывает GET /api/v1/topology/dot
func (h *Handler) topologyDOTHandler(w http.ResponseWriter, r *http.Request) {
	var req topologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.applyDefaults()

	topo, failure := h.buildTopologyForRequest(req)
	if failure != nil {
		h.writeError(w, failure.code, failure.msg)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	topo.ToDOT(w)
}

// topologyStatsHandler обрабатывает GET /api/v1/topology/stats
func (h *Handler) topologyStatsHandler(w http.ResponseWriter, r *http.Request) {
	var req topologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.applyDefaults()

	topo, failure := h.buildTopologyForRequest(req)
	if failure != nil {
		h.writeError(w, failure.code, failure.msg)
		return
	}

	// Статистика по типам устройств
	typeStats := make(map[string]int)
	for _, d := range topo.Devices {
		typeStats[string(d.Type)]++
	}

	// Статистика по confidence и source_type
	confidenceStats := make(map[string]int)
	sourceStats := make(map[string]int)
	for _, l := range topo.Links {
		confidenceStats[string(l.Confidence)]++
		sourceStats[string(l.SourceType)]++
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"device_count":     len(topo.Devices),
		"link_count":       len(topo.Links),
		"type_stats":       typeStats,
		"confidence_stats": confidenceStats,
		"source_stats":     sourceStats,
	})
}

// Вспомогательные функции

func deviceDisplayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func portLabel(p *topology.Port) string {
	if p == nil {
		return ""
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return ""
}
