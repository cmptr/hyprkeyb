package hyprland

import (
	"reflect"
	"testing"

	"github.com/cmptr/hyprkeyb/config"
)

func TestDecodeMods(t *testing.T) {
	tests := []struct {
		modmask uint
		want    string
	}{
		{0, ""},
		{1, "Shift"},
		{4, "Ctrl"},
		{8, "Alt"},
		{64, "Super"},
		{65, "Super + Shift"},
		{68, "Super + Ctrl"},
		{72, "Super + Alt"},
		{77, "Super + Ctrl + Alt + Shift"},
		{32, ""},   // unknown bit, silently ignored
		{96, "Super"}, // 64+32, unknown bit ignored
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := decodeMods(tt.modmask)
			if got != tt.want {
				t.Errorf("decodeMods(%d) = %q, want %q", tt.modmask, got, tt.want)
			}
		})
	}
}

func TestFormatKey(t *testing.T) {
	tests := []struct {
		mods string
		key  string
		want string
	}{
		{"Super", "V", "Super + V"},
		{"Super + Shift", "F", "Super + Shift + F"},
		{"", "Return", "Return"},
		{"Super", "XF86AudioRaiseVolume", "Super + XF86AudioRaiseVolume"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := formatKey(tt.mods, tt.key)
			if got != tt.want {
				t.Errorf("formatKey(%q, %q) = %q, want %q", tt.mods, tt.key, got, tt.want)
			}
		})
	}
}

func TestResolveName(t *testing.T) {
	tests := []struct {
		desc string
		disp string
		arg  string
		want string
	}{
		{"open terminal", "exec", "ghostty", "open terminal"},
		{"", "exec", "ghostty --class=x -e y", "exec: ghostty"},
		{"", "exec", "", "exec"},
		{"", "exec", "   ", "exec"},
		{"", "togglefloating", "", "togglefloating"},
		{"", "movefocus", "l", "movefocus: l"},
		{"", "workspace", "1", "workspace: 1"},
		// path-based exec: basename without extension
		{"", "exec", "~/.local/share/scripts/obsidian-quicknote.sh", "exec: obsidian-quicknote"},
		{"", "exec", "~/.config/rofi/clipboard/launcher.sh", "exec: launcher"},
		{"", "exec", "/usr/bin/firefox --new-window", "exec: firefox"},
		// path with no extension left unchanged as basename
		{"", "exec", "/usr/local/bin/mytool", "exec: mytool"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := resolveName(tt.desc, tt.disp, tt.arg)
			if got != tt.want {
				t.Errorf("resolveName(%q,%q,%q) = %q, want %q",
					tt.desc, tt.disp, tt.arg, got, tt.want)
			}
		})
	}
}

// fixture matches `hyprctl binds -j` structure
const testFixture = `[
  {
    "locked": false, "mouse": false, "release": false, "repeat": false,
    "longPress": false, "non_consuming": false, "has_description": false,
    "modmask": 64, "submap": "", "submap_universal": "false",
    "key": "Return", "keycode": 0, "catch_all": false,
    "description": "", "dispatcher": "exec", "arg": "ghostty --title=term"
  },
  {
    "locked": false, "mouse": false, "release": false, "repeat": false,
    "longPress": false, "non_consuming": false, "has_description": true,
    "modmask": 65, "submap": "", "submap_universal": "false",
    "key": "Q", "keycode": 0, "catch_all": false,
    "description": "close window", "dispatcher": "killactive", "arg": ""
  },
  {
    "locked": false, "mouse": false, "release": false, "repeat": false,
    "longPress": false, "non_consuming": false, "has_description": false,
    "modmask": 64, "submap": "resize", "submap_universal": "false",
    "key": "h", "keycode": 0, "catch_all": false,
    "description": "", "dispatcher": "resizeactive", "arg": "-10 0"
  }
]`

func TestParseBindsFromJSON(t *testing.T) {
	apps, err := parseBindsFromJSON([]byte(testFixture))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(apps) != 2 {
		t.Fatalf("expected 2 apps (Hyprland + Hyprland: resize), got %d", len(apps))
	}

	// "Hyprland" should be first
	if apps[0].Name != "Hyprland" {
		t.Errorf("first app name = %q, want %q", apps[0].Name, "Hyprland")
	}
	if len(apps[0].Keybinds) != 2 {
		t.Errorf("Hyprland keybinds count = %d, want 2", len(apps[0].Keybinds))
	}

	// verify name resolution: description takes priority
	wantBind := config.KeyBind{Name: "close window", Key: "Super + Shift + Q"}
	got := apps[0].Keybinds[1]
	if !reflect.DeepEqual(got, wantBind) {
		t.Errorf("bind = %+v, want %+v", got, wantBind)
	}

	// "Hyprland: resize" submap section
	if apps[1].Name != "Hyprland: resize" {
		t.Errorf("second app name = %q, want %q", apps[1].Name, "Hyprland: resize")
	}
}
