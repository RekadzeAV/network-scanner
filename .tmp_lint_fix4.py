import io

p = 'internal/gui/theme_manager_test.go'
t = io.open(p, 'r', encoding='utf-8', newline='').read()

old = """\ttm := NewThemeManager(nil)

\ttoggleCalled := false
\tui := NewThemeSwitcherUI(tm, func(mode ThemeMode) {
\t\ttoggleCalled = true
\t})

\tui.Toggle()
\tif !toggleCalled {
\t\tt.Error("expected toggle callback to be called")
\t}

\t// \u041f\u0440\u043e\u0432\u0435\u0440\u044f\u0435\u043c \u0447\u0442\u043e \u0442\u0435\u043c\u0430 \u0438\u0437\u043c\u0435\u043d\u0438\u043b\u0430\u0441\u044c
\t// \u0422\u0435\u043c\u0430 \u0434\u043e\u043b\u0436\u043d\u0430 \u043f\u0435\u0440\u0435\u043a\u043b\u044e\u0447\u0438\u0442\u044c\u0441\u044f \u0441 \u0442\u0435\u043a\u0443\u0449\u0435\u0439 \u043d\u0430 \u043f\u0440\u043e\u0442\u0438\u0432\u043e\u043f\u043e\u043b\u043e\u0436\u043d\u0443\u044e
\t_ = ThemeModeDark
\tif tm.GetThemeMode() == ThemeModeLight {
\t\t_ = ThemeModeLight
\t}
\tif tm.GetThemeMode() != before {
\t\t// \u0422\u0435\u043c\u0430 \u043f\u0435\u0440\u0435\u043a\u043b\u044e\u0447\u0438\u043b\u0430\u0441\u044c
\t}"""

new = """\ttm := NewThemeManager(nil)
\tbefore := tm.GetThemeMode()

\ttoggleCalled := false
\tui := NewThemeSwitcherUI(tm, func(mode ThemeMode) {
\t\ttoggleCalled = true
\t})

\tui.Toggle()
\tif !toggleCalled {
\t\tt.Error("expected toggle callback to be called")
\t}

\t// \u0422\u0435\u043c\u0430 \u0434\u043e\u043b\u0436\u043d\u0430 \u043f\u0435\u0440\u0435\u043a\u043b\u044e\u0447\u0438\u0442\u044c\u0441\u044f \u0441 \u0442\u0435\u043a\u0443\u0449\u0435\u0439 \u043d\u0430 \u043f\u0440\u043e\u0442\u0438\u0432\u043e\u043f\u043e\u043b\u043e\u0436\u043d\u0443\u044e
\tif tm.GetThemeMode() == before {
\t\tt.Errorf("expected theme mode to change from %v", before)
\t}"""

if '\r\n' in t:
    old = old.replace('\n', '\r\n')
    new = new.replace('\n', '\r\n')

if old in t:
    io.open(p, 'w', encoding='utf-8', newline='').write(t.replace(old, new))
    print('EDITED')
else:
    print('MISS')
