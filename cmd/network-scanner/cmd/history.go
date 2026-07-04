package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"network-scanner/internal/inventory"
)

// ExecuteHistory запускает команду истории
func ExecuteHistory(inventoryPath string, limit int, compareA, compareB string) error {
	store, err := inventory.Open(inventoryPath)
	if err != nil {
		return fmt.Errorf("open inventory: %w", err)
	}
	defer store.Close()

	if compareA != "" && compareB != "" {
		return executeCompare(store, compareA, compareB)
	}

	return executeList(store, limit)
}

func executeList(store *inventory.Store, limit int) error {
	history, _, err := store.GetScanHistory(limit)
	if err != nil {
		return fmt.Errorf("get scan history: %w", err)
	}

	if len(history) == 0 {
		fmt.Println("No scan history found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SCAN ID\tNETWORK\tHOSTS\tOS TYPES\tVENDORS\tDATE")
	fmt.Fprintln(w, "--------\t-------\t-----\t---------\t---------\t------")

	for _, entry := range history {
		osTypes := formatMap(entry.OSMap)
		vendors := formatMap(entry.VendorMap)
		fmt.Fprintf(w, "%s\t-\t%d\t%s\t%s\t%s\n",
			entry.ID[:min(8, len(entry.ID))],
			entry.HostCount,
			osTypes,
			vendors,
			entry.StartedAt.Format("2006-01-02 15:04"),
		)
	}

	w.Flush()
	return nil
}

func executeCompare(store *inventory.Store, scanIDA, scanIDB string) error {
	result, err := store.CompareSnapshotsByName(scanIDA, scanIDB)
	if err != nil {
		return fmt.Errorf("compare snapshots: %w", err)
	}

	fmt.Printf("Comparing %s vs %s\n\n", scanIDA[:min(8, len(scanIDA))], scanIDB[:min(8, len(scanIDB))])

	fmt.Printf("Total changes: %d\n\n", result.TotalDiff)

	if len(result.NewHosts) > 0 {
		fmt.Printf("New hosts (%d):\n", len(result.NewHosts))
		for _, h := range result.NewHosts {
			fmt.Printf("  + %s (%s)\n", h.IP, h.Hostname)
		}
		fmt.Println()
	}

	if len(result.RemovedHosts) > 0 {
		fmt.Printf("Removed hosts (%d):\n", len(result.RemovedHosts))
		for _, h := range result.RemovedHosts {
			fmt.Printf("  - %s (%s)\n", h.IP, h.Hostname)
		}
		fmt.Println()
	}

	if len(result.ChangedHosts) > 0 {
		fmt.Printf("Changed hosts (%d):\n", len(result.ChangedHosts))
		for _, c := range result.ChangedHosts {
			fmt.Printf("  ~ %s (%s): %s\n", c.IP, c.Hostname, joinStrings(c.ChangedIn))
		}
		fmt.Println()
	}

	if len(result.PortChanges) > 0 {
		fmt.Printf("Port changes (%d):\n", len(result.PortChanges))
		for _, p := range result.PortChanges {
			fmt.Printf("  %s:%d %s: %s -> %s\n", p.HostIP, p.Port, p.Protocol, p.ChangedFrom, p.ChangedTo)
		}
		fmt.Println()
	}

	return nil
}

func formatMap(m map[string]int) string {
	if len(m) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s:%d", k, v))
	}
	return joinStrings(parts)
}

func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return "-"
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += ", " + s
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


