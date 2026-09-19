import sys
from pathlib import Path

def patch_file(path, transformations):
    p = Path(path)
    t = p.read_text(encoding='utf-8')
    for old, new in transformations:
        if old in t:
            t = t.replace(old, new)
        else:
            print('NOT FOUND in', path, ':', repr(old[:40]))
    p.write_text(t, encoding='utf-8')
    print('PATCHED', path)

if __name__ == '__main__':
    patch_file('internal/gui/app_shortcuts.go', [
        ('canvas.AddShortcut(&desktop.CustomShortcut{
		Serialization: keys,
		Label:         label,
	}, func(_ *fyne.Shortcut) { h() })',
         'canvas.AddShortcut(&desktop.CustomShortcut{
		Keys:  keys,
		Label: label,
	}, func(shortcut fyne.Shortcut) { h() })'),
        ('fyne.NewDragOptions()', 'nil'),
    ])
    patch_file('internal/gui/ui_helpers.go', [
        ('fyne.NewDragOptions()', 'nil'),
    ])
    print('ALL DONE')
