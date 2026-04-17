package hyprland

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kencx/keyb/config"
)

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

type hyprBind struct {
	Modmask     uint   `json:"modmask"`
	Key         string `json:"key"`
	Dispatcher  string `json:"dispatcher"`
	Arg         string `json:"arg"`
	Description string `json:"description"`
	Submap      string `json:"submap"`
}

// ParseBinds queries the running Hyprland compositor and returns its
// keybinds as config.Apps grouped by submap.
func ParseBinds() (config.Apps, error) {
	out, err := exec.Command("hyprctl", "binds", "-j").Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl binds failed: %w", err)
	}
	return parseBindsFromJSON(out)
}

func parseBindsFromJSON(data []byte) (config.Apps, error) {
	var binds []hyprBind
	if err := json.Unmarshal(data, &binds); err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl output: %w", err)
	}

	groups := make(map[string][]config.KeyBind)
	for _, b := range binds {
		mods := decodeMods(b.Modmask)
		key := formatKey(mods, b.Key)
		name := resolveName(b.Description, b.Dispatcher, b.Arg)

		section := sectionName(b.Submap)
		groups[section] = append(groups[section], config.KeyBind{
			Name: name,
			Key:  key,
		})
	}

	apps := make(config.Apps, 0, len(groups))
	for section, keybinds := range groups {
		apps = append(apps, &config.App{
			Name:     section,
			Keybinds: keybinds,
		})
	}

	// "Hyprland" first, then submaps alphabetically
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Name == "Hyprland" {
			return true
		}
		if apps[j].Name == "Hyprland" {
			return false
		}
		return apps[i].Name < apps[j].Name
	})

	return apps, nil
}

func sectionName(submap string) string {
	if submap == "" {
		return "Hyprland"
	}
	return "Hyprland: " + submap
}

func resolveName(description, dispatcher, arg string) string {
	if description != "" {
		return description
	}
	if arg == "" {
		return dispatcher
	}
	if dispatcher == "exec" {
		parts := strings.Fields(arg)
		if len(parts) > 0 {
			cmd := parts[0]
			if strings.Contains(cmd, "/") {
				base := filepath.Base(cmd)
				cmd = strings.TrimSuffix(base, filepath.Ext(base))
			}
			return "exec: " + cmd
		}
		return dispatcher
	}
	return dispatcher + ": " + arg
}
