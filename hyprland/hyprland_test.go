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
