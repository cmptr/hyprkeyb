package ghostty

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/cmptr/hyprkeyb/config"
)

// actionRe matches the last =<lowercase-word> in a binding string.
// This handles keys that contain = (e.g. ctrl+=).
var actionRe = regexp.MustCompile(`=([a-z][^=]*)$`)

// ParseBinds runs `ghostty +list-keybinds` and returns keybinds as config.Apps.
func ParseBinds() (config.Apps, error) {
	out, err := exec.Command("ghostty", "+list-keybinds").Output()
	if err != nil {
		return nil, fmt.Errorf("ghostty +list-keybinds failed: %w", err)
	}
	return parseBindsFromText(out)
}

func parseBindsFromText(data []byte) (config.Apps, error) {
	var keybinds []config.KeyBind

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "keybind = ") {
			continue
		}
		binding := strings.TrimPrefix(line, "keybind = ")

		loc := actionRe.FindStringSubmatchIndex(binding)
		if loc == nil {
			continue
		}
		key := binding[:loc[0]]
		action := binding[loc[2]:loc[3]]

		keybinds = append(keybinds, config.KeyBind{
			Name: actionToName(action),
			Key:  key,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan ghostty output: %w", err)
	}

	return config.Apps{
		&config.App{
			Name:     "Ghostty",
			Keybinds: keybinds,
		},
	}, nil
}

func actionToName(action string) string {
	base, param, hasParam := strings.Cut(action, ":")

	parts := strings.Split(base, "_")
	for i, p := range parts {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}
	name := strings.Join(parts, " ")

	if hasParam && param != "" {
		name += " (" + param + ")"
	}
	return name
}
