package hyprland

import "strings"

// X11 modifier bit masks
const (
	modShift = 1
	modCtrl  = 4
	modAlt   = 8
	modSuper = 64
	modMod5  = 128
)

var modOrder = []struct {
	bit  uint
	name string
}{
	{modSuper, "Super"},
	{modCtrl, "Ctrl"},
	{modAlt, "Alt"},
	{modShift, "Shift"},
	{modMod5, "Mod5"},
}

func decodeMods(modmask uint) string {
	var parts []string
	for _, m := range modOrder {
		if modmask&m.bit != 0 {
			parts = append(parts, m.name)
		}
	}
	return strings.Join(parts, " + ")
}
