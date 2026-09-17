"""Пакетные lint-фиксы: точечные замены с проверкой вхождения."""
import io, re, sys

def edit(path, pairs, regex_pairs=(), count_hint=None):
    with io.open(path, 'r', encoding='utf-8', newline='') as f:
        t = f.read()
    orig = t
    for old, new in pairs:
        if old not in t:
            print(f"MISS {path}: {old[:60]!r}")
        t = t.replace(old, new)
    for pat, new in regex_pairs:
        t2, n = re.subn(pat, new, t, flags=re.S | re.M)
        if n == 0:
            print(f"MISSRE {path}: {pat[:60]!r}")
        t = t2
    if t != orig:
        with io.open(path, 'w', encoding='utf-8', newline='') as f:
            f.write(t)
        print(f"EDITED {path}")

# 1) scan.go: context import
edit('cmd/network-scanner/cmd/scan.go', [
    ('import (\n\t"fmt"', 'import (\n\t"context"\n\t"fmt"'),
    ('import (\r\n\t"fmt"', 'import (\r\n\t"context"\r\n\t"fmt"'),
])

# 2) remote_exec.go: nil ctx -> context.TODO
with io.open('cmd/network-scanner/cmd/remote_exec.go', 'r', encoding='utf-8', newline='') as f:
    re_t = f.read()
print("remote_exec imports context:", '"context"' in re_t)

# 3) results_table_view.go type assertions
edit('internal/gui/results_table_view.go', [
    ('l := obj.(*widget.Label)', 'l, _ := obj.(*widget.Label)\n\tif l == nil {\n\t\treturn\n\t}'),
])

# 4) scan_profile.go: profileName
edit('internal/gui/controller/scan_profile.go', [
    ('\tprofileName := "стандарт"\n\tswitch {', '\tprofileName := "стандарт"\n\tswitch {'),
])

# 5) gosimple S1009 tests
edit('internal/gui/model_edge_cases_test.go', [
    ('if results[0].Ports != nil && len(results[0].Ports) != 0 {', 'if len(results[0].Ports) != 0 {'),
])
edit('internal/gui/results_view_test.go', [
    ('if filters != nil && len(filters) != 0 {', 'if len(filters) != 0 {'),
])
edit('internal/scanner/scanner_coverage_extend_test.go', [
    ('if results != nil && len(results) != 0 {', 'if len(results) != 0 {'),
])
edit('internal/security/security_integration_test.go', [
    ('if v != nil && len(v) == 0 {', 'if len(v) == 0 {'),
])

# 6) nettools SA4031: remove impossible nil checks
edit('internal/nettools/nettools_integration_test.go', [
    ('\tresult := &DNSResult{}\n\tif result == nil {\n\t\tt.Error("expected non-nil DNSResult")\n\t}',
     '\tresult := &DNSResult{}\n\tif len(result.Addresses) != 0 {\n\t\tt.Errorf("expected empty addresses, got %v", result.Addresses)\n\t}'),
    ('\tresult := &PingResult{}\n\tif result == nil {\n\t\tt.Error("expected non-nil PingResult")\n\t}',
     '\tresult := &PingResult{}\n\tif result.Sent != 0 {\n\t\tt.Errorf("expected zero sent, got %d", result.Sent)\n\t}'),
    ('\tresult := &TracerouteResult{}\n\tif result == nil {\n\t\tt.Error("expected non-nil TracerouteResult")\n\t}',
     '\tresult := &TracerouteResult{}\n\tif len(result.Hops) != 0 {\n\t\tt.Errorf("expected empty hops, got %v", result.Hops)\n\t}'),
])

print("part1 done")
