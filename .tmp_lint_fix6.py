"""Последние lint-фиксы (партия 4)."""
import io

def read(path):
    with io.open(path, 'r', encoding='utf-8', newline='') as f:
        return f.read()

def write(path, t):
    with io.open(path, 'w', encoding='utf-8', newline='') as f:
        f.write(t)

def fix(path, pairs):
    t = read(path)
    for old, new in pairs:
        if old in t:
            t = t.replace(old, new)
        elif old.replace('\n', '\r\n') in t:
            t = t.replace(old.replace('\n', '\r\n'), new.replace('\n', '\r\n'))
        else:
            print(f"MISS {path}: {old[:55]!r}")
    write(path, t)

fix('internal/cve/coverage_test.go', [
    ('\tif len(matches) < 0 {\n\t\tt.Fatal("matches should not be negative")\n\t}',
     '\tt.Logf("matches: %d", len(matches))'),
])

fix('internal/scanner/scanner_integration_test.go', [
    ('\tvar ns *NetworkScanner\n\tif ns == nil {\n\t\t// Expected - nil scanner\n\t}',
     '\tvar ns *NetworkScanner\n\tif ns != nil {\n\t\tt.Fatal("expected nil scanner")\n\t}'),
])

fix('internal/scanner/plugin/builtin_probes.go', [
    ('\tn, err := conn.Read(buf)\n\tif err != nil {\n\t\t// \u041c\u043e\u0436\u0435\u0442 \u0431\u044b\u0442\u044c timeout \u2014 \u044d\u0442\u043e \u043d\u043e\u0440\u043c\u0430\u043b\u044c\u043d\u043e\n\t}',
     '\tn, err := conn.Read(buf)\n\tif err != nil {\n\t\t// timeout \u2014 \u043d\u043e\u0440\u043c\u0430\u043b\u044c\u043d\u043e, banner \u043c\u043e\u0436\u0435\u0442 \u043e\u0442\u0441\u0443\u0442\u0441\u0442\u0432\u043e\u0432\u0430\u0442\u044c\n\t\tn = 0\n\t}'),
])

fix('internal/gui/security_view.go', [
    ('\t\t\tl := obj.(*widget.Label)\n\t\t\tif id.Row == 0 {',
     '\t\t\tl, ok := obj.(*widget.Label)\n\t\t\tif !ok || l == nil {\n\t\t\t\treturn\n\t\t\t}\n\t\t\tif id.Row == 0 {'),
])

fix('internal/gui/mobile_layout.go', [
    ('\tmainTabs           *container.AppTabs\n\ttabs               []*container.TabItem\n}',
     '\tmainTabs           *container.AppTabs\n}'),
])

print("PART4 DONE")
