"""Пакетные lint-фиксы (партия 2)."""
import io

def read(path):
    with io.open(path, 'r', encoding='utf-8', newline='') as f:
        return f.read()

def write(path, t):
    with io.open(path, 'w', encoding='utf-8', newline='') as f:
        f.write(t)

def fix(path, pairs):
    t = read(path)
    ok = True
    for old, new in pairs:
        if old in t:
            t = t.replace(old, new)
        elif old.replace('\n', '\r\n') in t:
            t = t.replace(old.replace('\n', '\r\n'), new.replace('\n', '\r\n'))
        else:
            print(f"MISS {path}: {old[:55]!r}")
            ok = False
    write(path, t)
    return ok

# --- remote_exec.go: context.TODO ---
fix('cmd/network-scanner/cmd/remote_exec.go', [
    ('import (\n\t"fmt"', 'import (\n\t"context"\n\t"fmt"'),
    ('remoteExecService.DryRun(nil, req)', 'remoteExecService.DryRun(context.TODO(), req)'),
    ('remoteExecService.Execute(nil, req)', 'remoteExecService.Execute(context.TODO(), req)'),
])

# --- schema_validate.go: S1011 ---
fix('internal/topology/schema_validate.go', [
    ('\t\tfor _, e := range gmlErrors {\n\t\t\tcheck.Errors = append(check.Errors, e)\n\t\t}',
     '\t\tcheck.Errors = append(check.Errors, gmlErrors...)'),
])

# --- telemetry.go: empty if ---
fix('internal/telemetry/telemetry.go', [
    ('\tif resp.StatusCode >= 200 && resp.StatusCode < 300 {\n\t\t// Успешно отправлено\n\t}\n', ''),
])

# --- errors_test.go: empty branch ---
fix('internal/errors/errors_test.go', [
    ('\tif !errors.Is(wrapped, ErrNotFound) {\n\t\t// Should still be able to unwrap\n\t}',
     '\tif errors.Is(wrapped, ErrNotFound) {\n\t\tt.Error("wrapped error unexpectedly matches ErrNotFound")\n\t}'),
])

# --- alerting_integration_test.go ---
fix('internal/alerting/alerting_integration_test.go', [
    ('\tif len(highAlerts) < 0 {\n\t\tt.Error("expected non-negative HIGH alerts count")\n\t}\n'
     '\tif len(mediumAlerts) < 0 {\n\t\tt.Error("expected non-negative MEDIUM alerts count")\n\t}\n'
     '\tif len(lowAlerts) < 0 {\n\t\tt.Error("expected non-negative LOW alerts count")\n\t}',
     '\tt.Logf("alerts: HIGH=%d MEDIUM=%d LOW=%d", len(highAlerts), len(mediumAlerts), len(lowAlerts))'),
    ('\terr := handler.OnAlert(alert)\n\tif err == nil {\n\t\t// On some systems, this might succeed or fail differently\n\t\t// Just verify the alert structure is valid\n\t}',
     '\terr := handler.OnAlert(alert)\n\tif err != nil {\n\t\tt.Logf("OnAlert may depend on environment: %v", err)\n\t}'),
])

# --- api_coverage_extended_test.go: "" == "" ---
fix('internal/api/api_coverage_extended_test.go', [
    ('\t// Симулируем пустые ID — код проходит\n\tif "" == "" {\n\t}',
     '\t// Симулируем пустые ID — код проходит'),
])

# --- theme_manager_test.go: tautology ---
fix('internal/gui/theme_manager_test.go', [
    ('\tif tm.GetThemeMode() == tm.GetThemeMode() {\n\t\t// Тема переключилась\n\t}',
     '\tif tm.GetThemeMode() != before {\n\t\t// Тема переключилась\n\t}'),
])

# --- osdetect example_test.go ---
fix('internal/osdetect/example_test.go', [
    ('osName, conf, reason = osdetect.GuessFromHostAndPorts("unknown-device", []int{8080}, true)',
     'osName, _, _ = osdetect.GuessFromHostAndPorts("unknown-device", []int{8080}, true)'),
])

# --- scan_profile.go: profileName ---
fix('internal/gui/controller/scan_profile.go', [
    ('\tprofileName := "стандарт"\n\tswitch {', '\tvar profileName string\n\tswitch {'),
])

# --- tools_controller.go: S1021 ---
fix('internal/gui/controller/tools_controller.go', [
    ('\t\tvar output string\n\t\tvar err error\n\n\t\terr = errors.ExecuteWithRetry(',
     '\t\tvar output string\n\n\t\terr := errors.ExecuteWithRetry('),
])

# --- scan_controller.go: ineffectual threads ---
fix('internal/gui/controller/scan_controller.go', [
    ('\t\tif threads > maxScanThreadsGUI {\n\t\t\tthreads = maxScanThreadsGUI\n\t\t\tc.ui.ThreadsEntry.SetText(strconv.Itoa(maxScanThreadsGUI))',
     '\t\tif threads > maxScanThreadsGUI {\n\t\t\tc.ui.ThreadsEntry.SetText(strconv.Itoa(maxScanThreadsGUI))'),
])

# --- scanner.go: dead flags ---
fix('internal/scanner/scanner.go', [
    ('\tcancelledBeforeLaunch := false\n', ''),
    ('\t\tif ns.ctx.Err() != nil {\n\t\t\tif !cancelledBeforeLaunch {\n\t\t\t\tatomic.AddInt64(&ns.tcpCancelBefore, 1)\n\t\t\t\tcancelledBeforeLaunch = true\n\t\t\t}\n\t\t\tbreak\n\t\t}',
     '\t\tif ns.ctx.Err() != nil {\n\t\t\tatomic.AddInt64(&ns.tcpCancelBefore, 1)\n\t\t\tbreak\n\t\t}'),
    ('\tudpScanCancelled := false\n udpPortLoop:', ' udpPortLoop:'),
    ('\t\tcase <-ns.ctx.Done():\n\t\t\tatomic.AddInt64(&ns.udpCancelHosts, 1)\n\t\t\tudpScanCancelled = true\n\t\t\tbreak udpPortLoop\n\t\tdefault:\n\t\t}\n\t\tif udpScanCancelled {\n\t\t\tbreak udpPortLoop\n\t\t}',
     '\t\tcase <-ns.ctx.Done():\n\t\t\tatomic.AddInt64(&ns.udpCancelHosts, 1)\n\t\t\tbreak udpPortLoop\n\t\tdefault:\n\t\t}'),
])

# --- split_persist_extended_test.go: S1021 ---
fix('internal/gui/split_persist_extended_test.go', [
    ('\tvar lastVal *float64\n\tlastVal = new(float64)', '\tlastVal := new(float64)'),
])

# --- nettools SA5011: t.Fatal on nil ---
fix('internal/nettools/nettools_integration_test.go', [
    ('\t\tif result == nil {\n\t\t\tt.Error("expected non-nil result")\n\t\t}\n\t\tif result.Stats.Sent < 0 {',
     '\t\tif result == nil {\n\t\t\tt.Fatal("expected non-nil result")\n\t\t}\n\t\tif result.Stats.Sent < 0 {'),
])

# --- deprecated fyne APIs ---
fix('internal/gui/results_settings_ui.go', [
    ('a.myWindow.Clipboard().SetContent(text)', 'a.myApp.Clipboard().SetContent(text)'),
])
fix('internal/gui/controller/scan_controller.go', [
    ('c.ui.Window.Clipboard().SetContent(diagnosticsText)', 'fyne.CurrentApp().Clipboard().SetContent(diagnosticsText)'),
])
fix('internal/gui/controller/topology_controller.go', [
    ('window.Clipboard().SetContent(reportText)', 'fyne.CurrentApp().Clipboard().SetContent(reportText)'),
])
fix('internal/gui/init_ui.go', [
    ('container.NewMax(', 'container.NewStack('),
])
fix('internal/gui/results_view_render.go', [
    ('a.mainTabs.SelectTabIndex(', 'a.mainTabs.SelectIndex('),
    ('container.NewMax(', 'container.NewStack('),
])
fix('internal/gui/scan_ui.go', [
    ('a.filtersInfoLabel.Wrapping = fyne.TextTruncate', 'a.filtersInfoLabel.Truncation = fyne.TextTruncateClip'),
    ('a.resultsPerfLabel.Wrapping = fyne.TextTruncate', 'a.resultsPerfLabel.Truncation = fyne.TextTruncateClip'),
])
fix('internal/gui/results_table_view.go', [
    ('container.NewMax(', 'container.NewStack('),
])

print("PART2 DONE")
