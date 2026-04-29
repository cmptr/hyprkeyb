package ghostty

import (
	"testing"
)

func TestParseBindsFromText(t *testing.T) {
	input := []byte(`
keybind = ctrl+shift+,=reload_config
keybind = ctrl+,=open_config
keybind = copy=copy_to_clipboard:mixed
keybind = ctrl+shift+c=copy_to_clipboard:mixed
keybind = ctrl+==increase_font_size:1
keybind = ctrl++=increase_font_size:1
keybind = ctrl+-=decrease_font_size:1
keybind = ctrl+0=reset_font_size
keybind = ctrl+shift+t=new_tab
keybind = alt+1=goto_tab:1
`)

	apps, err := parseBindsFromText(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 {
		t.Fatalf("want 1 app, got %d", len(apps))
	}
	app := apps[0]
	if app.Name != "Ghostty" {
		t.Errorf("want name Ghostty, got %q", app.Name)
	}
	if len(app.Keybinds) != 10 {
		t.Errorf("want 10 keybinds, got %d", len(app.Keybinds))
	}

	cases := []struct {
		idx  int
		name string
		key  string
	}{
		{0, "Reload Config", "ctrl+shift+,"},
		{4, "Increase Font Size (1)", "ctrl+="},
		{5, "Increase Font Size (1)", "ctrl++"},
		{9, "Goto Tab (1)", "alt+1"},
	}
	for _, c := range cases {
		kb := app.Keybinds[c.idx]
		if kb.Name != c.name {
			t.Errorf("[%d] name: want %q, got %q", c.idx, c.name, kb.Name)
		}
		if kb.Key != c.key {
			t.Errorf("[%d] key: want %q, got %q", c.idx, c.key, kb.Key)
		}
	}
}

func TestActionToName(t *testing.T) {
	cases := []struct {
		action string
		want   string
	}{
		{"reload_config", "Reload Config"},
		{"copy_to_clipboard:mixed", "Copy To Clipboard (mixed)"},
		{"increase_font_size:1", "Increase Font Size (1)"},
		{"reset_font_size", "Reset Font Size"},
		{"goto_tab:1", "Goto Tab (1)"},
		{"new_tab", "New Tab"},
		{"quit", "Quit"},
	}
	for _, c := range cases {
		got := actionToName(c.action)
		if got != c.want {
			t.Errorf("actionToName(%q) = %q, want %q", c.action, got, c.want)
		}
	}
}
