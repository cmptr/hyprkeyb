package hyprland

import (
	"testing"
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
