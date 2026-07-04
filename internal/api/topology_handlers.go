package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"github.com/gorilla/mux"
)

// topologyBuildHandler РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ POST /api/v1/topology/build
func (h *Handler) topologyBuildHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SnapshotID    string `json:"snapshot_id"`
		SNMPEnabled   bool   `json:"snmp_enabled"`
		SNMPCommunity string `json:"snmp_community"`
		SNMPTimeout   int    `json:"snmp_timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SNMPTimeout <= 0 {
		req.SNMPTimeout = 2
	}
	if req.SNMPCommunity == "" {
		req.SNMPCommunity = "public"
	}

	// Р—Р°РіСЂСѓР¶Р°РµРј СЃРЅР°РїС€РѕС‚
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	var hosts []scanner.Result
	if req.SnapshotID != "" {
		snap, err := store.LoadSnapshot(req.SnapshotID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot not found: %v", err))
			return
		}
		hosts = snap.Hosts
	} else {
		// Р‘РµСЂС‘Рј РїРѕСЃР»РµРґРЅРёР№ СЃРЅР°РїС€РѕС‚
		snapshots, err := store.ListSnapshots(1)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err))
			return
		}
		if len(snapshots) == 0 {
			h.writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "no snapshots found",
			})
			return
		}
		lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err))
			return
		}
		hosts = lastSnap.Hosts
	}

	// SNMP РѕРїСЂРѕСЃ
	var snmpData map[string]*topology.Device
	if req.SNMPEnabled {
		fmt.Printf("SNMP РѕРїСЂРѕСЃ РґР»СЏ С‚РѕРїРѕР»РѕРіРёРё: %d СѓСЃС‚СЂРѕР№СЃС‚РІ\n", len(hosts))
		snmpData, _, err = snmpcollector.CollectWithReport(hosts, []string{req.SNMPCommunity}, req.SNMPTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		}
	}

	// РЎС‚СЂРѕРёРј С‚РѕРїРѕР»РѕРіРёСЋ
	topo, err := topology.BuildTopology(hosts, snmpData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("build topology: %v", err))
		return
	}

	// РљРѕРЅРІРµСЂС‚РёСЂСѓРµРј РІ JSON
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

// topologyExportHandler РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ POST /api/v1/topology/export/{format}
func (h *Handler) topologyExportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]

	if format != "json" && format != "dot" && format != "graphml" {
		h.writeError(w, http.StatusBadRequest, "format must be json, dot, or graphml")
		return
	}

	var req struct {
		SnapshotID    string `json:"snapshot_id"`
		SNMPEnabled   bool   `json:"snmp_enabled"`
		SNMPCommunity string `json:"snmp_community"`
		SNMPTimeout   int    `json:"snmp_timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SNMPTimeout <= 0 {
		req.SNMPTimeout = 2
	}
	if req.SNMPCommunity == "" {
		req.SNMPCommunity = "public"
	}

	// Р—Р°РіСЂСѓР¶Р°РµРј СЃРЅР°РїС€РѕС‚
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	var hosts []scanner.Result
	if req.SnapshotID != "" {
		snap, err := store.LoadSnapshot(req.SnapshotID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot not found: %v", err))
			return
		}
		hosts = snap.Hosts
	} else {
		snapshots, err := store.ListSnapshots(1)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err))
			return
		}
		if len(snapshots) == 0 {
			h.writeError(w, http.StatusNotFound, "no snapshots found")
			return
		}
		lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err))
			return
		}
		hosts = lastSnap.Hosts
	}

	// SNMP РѕРїСЂРѕСЃ
	var snmpData map[string]*topology.Device
	if req.SNMPEnabled {
		snmpData, _, err = snmpcollector.CollectWithReport(hosts, []string{req.SNMPCommunity}, req.SNMPTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		}
	}

	// РЎС‚СЂРѕРёРј С‚РѕРїРѕР»РѕРіРёСЋ
	topo, err := topology.BuildTopology(hosts, snmpData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("build topology: %v", err))
		return
	}

	// Р­РєСЃРїРѕСЂС‚РёСЂСѓРµРј
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
		w.Header().Set("Content-Type", "application/xml")
		tmpFile, err := os.CreateTemp("", "topology-*.graphml")
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("create temp file: %v", err))
			return
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)
		tmpFile.Close()

		if err := topo.SaveGraphML(tmpPath); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("save graphml: %v", err))
			return
		}

		data, err := ioutil.ReadFile(tmpPath)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("read temp file: %v", err))
			return
		}
		w.Write(data)
	}
}

// topologyDOTHandler РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ GET /api/v1/topology/dot
func (h *Handler) topologyDOTHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SnapshotID    string `json:"snapshot_id"`
		SNMPEnabled   bool   `json:"snmp_enabled"`
		SNMPCommunity string `json:"snmp_community"`
		SNMPTimeout   int    `json:"snmp_timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SNMPTimeout <= 0 {
		req.SNMPTimeout = 2
	}
	if req.SNMPCommunity == "" {
		req.SNMPCommunity = "public"
	}

	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	var hosts []scanner.Result
	if req.SnapshotID != "" {
		snap, err := store.LoadSnapshot(req.SnapshotID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot not found: %v", err))
			return
		}
		hosts = snap.Hosts
	} else {
		snapshots, err := store.ListSnapshots(1)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err))
			return
		}
		if len(snapshots) == 0 {
			h.writeError(w, http.StatusNotFound, "no snapshots found")
			return
		}
		lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err))
			return
		}
		hosts = lastSnap.Hosts
	}

	var snmpData map[string]*topology.Device
	if req.SNMPEnabled {
		snmpData, _, err = snmpcollector.CollectWithReport(hosts, []string{req.SNMPCommunity}, req.SNMPTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		}
	}

	topo, err := topology.BuildTopology(hosts, snmpData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("build topology: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	topo.ToDOT(w)
}

// topologyStatsHandler РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ GET /api/v1/topology/stats
func (h *Handler) topologyStatsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SnapshotID    string `json:"snapshot_id"`
		SNMPEnabled   bool   `json:"snmp_enabled"`
		SNMPCommunity string `json:"snmp_community"`
		SNMPTimeout   int    `json:"snmp_timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SNMPTimeout <= 0 {
		req.SNMPTimeout = 2
	}
	if req.SNMPCommunity == "" {
		req.SNMPCommunity = "public"
	}

	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	var hosts []scanner.Result
	if req.SnapshotID != "" {
		snap, err := store.LoadSnapshot(req.SnapshotID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot not found: %v", err))
			return
		}
		hosts = snap.Hosts
	} else {
		snapshots, err := store.ListSnapshots(1)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err))
			return
		}
		if len(snapshots) == 0 {
			h.writeError(w, http.StatusNotFound, "no snapshots found")
			return
		}
		lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err))
			return
		}
		hosts = lastSnap.Hosts
	}

	var snmpData map[string]*topology.Device
	if req.SNMPEnabled {
		snmpData, _, err = snmpcollector.CollectWithReport(hosts, []string{req.SNMPCommunity}, req.SNMPTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		}
	}

	topo, err := topology.BuildTopology(hosts, snmpData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("build topology: %v", err))
		return
	}

	// РЎС‚Р°С‚РёСЃС‚РёРєР° РїРѕ С‚РёРїР°Рј СѓСЃС‚СЂРѕР№СЃС‚РІ
	typeStats := make(map[string]int)
	for _, d := range topo.Devices {
		typeStats[string(d.Type)]++
	}

	// РЎС‚Р°С‚РёСЃС‚РёРєР° РїРѕ confidence
	confidenceStats := make(map[string]int)
	for _, l := range topo.Links {
		confidenceStats[string(l.Confidence)]++
	}

	// РЎС‚Р°С‚РёСЃС‚РёРєР° РїРѕ source_type
	sourceStats := make(map[string]int)
	for _, l := range topo.Links {
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

// Р’СЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹Рµ С„СѓРЅРєС†РёРё

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

