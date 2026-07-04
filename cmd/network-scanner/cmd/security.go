package cmd

import (
	"context"
	"fmt"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
)

// RunSecurity запускает анализ безопасности
func RunSecurity(cfg builder.Config, results []contracts.ScanResult) error {
	container := builder.NewContainer(cfg)
	securityService := container.GetSecurity()

	report, err := securityService.AnalyzeRun(context.Background(), results)
	if err != nil {
		return fmt.Errorf("анализ безопасности: %w", err)
	}

	fmt.Println("=== Security Report ===")
	fmt.Printf("Security Score: %d/100\n", report.Score)
	fmt.Printf("Port Audit Findings: %d\n", len(report.PortAudit))
	fmt.Printf("Risk Signature Findings: %d\n", len(report.RiskSig))

	if len(report.PortAudit) > 0 {
		fmt.Println("\n--- Port Audit ---")
		for _, f := range report.PortAudit {
			fmt.Printf("[%s] %s (host: %s)\n", f.Severity, f.Title, f.Host)
			if f.Recommendation != "" {
				fmt.Printf("  Recommendation: %s\n", f.Recommendation)
			}
		}
	}

	if len(report.RiskSig) > 0 {
		fmt.Println("\n--- Risk Signatures ---")
		for _, f := range report.RiskSig {
			fmt.Printf("[%s] %s (host: %s)\n", f.Severity, f.Title, f.Host)
			if f.Recommendation != "" {
				fmt.Printf("  Recommendation: %s\n", f.Recommendation)
			}
		}
	}

	return nil
}

// RunTopology строит топологию сети
func RunTopology(cfg builder.Config, results []contracts.ScanResult, snmpCommunity string, snmpTimeout int) error {
	container := builder.NewContainer(cfg)
	topologyService := container.GetTopology()

	opts := contracts.TopologyOptions{
		SNMPEnabled: snmpCommunity != "",
		Community:   snmpCommunity,
		Timeout:     time.Duration(snmpTimeout) * time.Second,
	}

	topo, err := topologyService.Build(context.Background(), results, opts)
	if err != nil {
		return fmt.Errorf("построение топологии: %w", err)
	}

	fmt.Printf("Топология построена: %d устройств, %d связей\n", len(topo.Devices), len(topo.Links))
	for _, dev := range topo.Devices {
		fmt.Printf("- %s (%s) type=%s\n", dev.IP, dev.Hostname, dev.Type)
	}
	for _, link := range topo.Links {
		fmt.Printf("  Link: %s -> %s (confidence: %s)\n",
			link.Source.IP, link.Target.IP, link.Confidence)
	}

	return nil
}

// RunRemoteExec выполняет удалённую команду
func RunRemoteExec(cfg builder.Config, transport, target, user, pass, command string, dryRun bool) error {
	container := builder.NewContainer(cfg)
	remoteExecService := container.GetRemoteExec()

	req := contracts.RemoteExecRequest{
		Transport: transport,
		Target:    target,
		User:      user,
		Password:  pass,
		Command:   command,
		DryRun:    dryRun,
		Timeout:   15 * time.Second,
	}

	if dryRun {
		fmt.Println("=== Dry Run ===")
		if err := remoteExecService.DryRun(context.Background(), req); err != nil {
			return fmt.Errorf("dry run failed: %w", err)
		}
		fmt.Println("Policy check passed")
		return nil
	}

	fmt.Println("=== Remote Exec ===")
	res, err := remoteExecService.Execute(context.Background(), req)
	if err != nil {
		return fmt.Errorf("remote exec failed: %w", err)
	}

	if res.Success {
		fmt.Printf("Success\nOutput:\n%s\n", res.Output)
	} else {
		fmt.Println("Failed")
	}

	return nil
}

// RunInventorySave сохраняет снапшот инвентаризации
func RunInventorySave(cfg builder.Config, results []contracts.ScanResult, id string) error {
	container := builder.NewContainer(cfg)
	inventoryService := container.GetInventory()

	if err := inventoryService.SaveSnapshot(context.Background(), id, results); err != nil {
		return fmt.Errorf("сохранение снапшота: %w", err)
	}

	fmt.Printf("Inventory snapshot сохранён: id=%s hosts=%d\n", id, len(results))
	return nil
}

// RunInventoryDiff сравнивает два снапшота
func RunInventoryDiff(cfg builder.Config, idA, idB string) error {
	container := builder.NewContainer(cfg)
	inventoryService := container.GetInventory()

	diff, err := inventoryService.Diff(context.Background(), idA, idB)
	if err != nil {
		return fmt.Errorf("inventory diff: %w", err)
	}

	fmt.Printf("Inventory diff: %s -> %s\n", diff.ScanIDA, diff.ScanIDB)
	fmt.Printf("- New: %d\n", len(diff.New))
	fmt.Printf("- Missing: %d\n", len(diff.Missing))
	fmt.Printf("- Changed: %d\n", len(diff.Changed))

	for _, c := range diff.Changed {
		fmt.Printf("  • %s fields=%s\n", c.Key, c.ChangedField)
	}

	return nil
}

// RunInventoryList показывает список снапшотов
func RunInventoryList(cfg builder.Config, limit int) error {
	container := builder.NewContainer(cfg)
	inventoryService := container.GetInventory()

	snapshots, err := inventoryService.ListSnapshots(context.Background(), limit)
	if err != nil {
		return fmt.Errorf("список снапшотов: %w", err)
	}

	fmt.Printf("Inventory snapshots (limit=%d):\n", limit)
	for _, snap := range snapshots {
		fmt.Printf("- %s at %s\n", snap.ID, snap.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// ConvertToInternalResults конвертирует contracts.ScanResult в scanner.Result
func ConvertToInternalResults(results []contracts.ScanResult) []scanner.Result {
	out := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}
		out = append(out, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}
	return out
}
