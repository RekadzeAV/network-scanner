import io, re

def fix(path, pairs):
    with io.open(path, 'r', encoding='utf-8', newline='') as f:
        t = f.read()
    for old, new in pairs:
        if old not in t:
            # try alternate line endings
            if '\r\n' in t:
                old2 = old.replace('\n', '\r\n')
                new2 = new.replace('\n', '\r\n')
            else:
                old2 = old.replace('\r\n', '\n')
                new2 = new.replace('\r\n', '\n')
            if old2 in t:
                t = t.replace(old2, new2)
                print(f"EDITED(crlf-alt) {path}")
            else:
                print(f"MISS {path}: {old[:50]!r}")
        else:
            t = t.replace(old, new)
            print(f"EDITED {path}")
    with io.open(path, 'w', encoding='utf-8', newline='') as f:
        f.write(t)

fix('internal/nettools/nettools_integration_test.go', [
    ('if len(result.Addresses) != 0 {\n\t\tt.Errorf("expected empty addresses, got %v", result.Addresses)\n\t}',
     'if len(result.ForwardIPs) != 0 {\n\t\tt.Errorf("expected empty forward IPs, got %v", result.ForwardIPs)\n\t}'),
    ('if result.Sent != 0 {\n\t\tt.Errorf("expected zero sent, got %d", result.Sent)\n\t}',
     'if result.Stats.Sent != 0 {\n\t\tt.Errorf("expected zero sent, got %d", result.Stats.Sent)\n\t}'),
    ('\tinfo := map[string]string{}\n\tif info == nil {\n\t\tt.Error("expected non-nil map")\n\t}',
     '\tinfo := map[string]string{}\n\tif len(info) != 0 {\n\t\tt.Errorf("expected empty map, got %v", info)\n\t}'),
])

fix('internal/security/security_integration_test.go', [
    ('if report.PortAudit != nil && len(report.PortAudit) != 0 {',
     'if len(report.PortAudit) != 0 {'),
])

print("done")
