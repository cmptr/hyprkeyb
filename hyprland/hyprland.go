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

func formatKey(mods, key string) string {
	if mods == "" {
		return key
	}
	return mods + " + " + key
}

func resolveName(description, dispatcher, arg string) string {
	if description != "" {
		return description
	}
	if arg == "" {
		return dispatcher
	}
	if dispatcher == "exec" {
		// show only first word of command for readability
		parts := strings.Fields(arg)
		if len(parts) > 0 {
			return "exec: " + parts[0]
		}
		return dispatcher
	}
	return dispatcher + ": " + arg
}
