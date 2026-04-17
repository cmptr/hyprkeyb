# Hyprland Auto-Detection & Nix Dev Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `-H` flag to keyb that auto-detects Hyprland keybinds via `hyprctl binds -j` and displays them alongside any manually configured binds; also add a Nix flake and `.envrc` for a reproducible dev environment.

**Architecture:** A new self-contained `hyprland/` package exposes `ParseBinds() (config.Apps, error)` which shells out to `hyprctl binds -j`, decodes X11 modmasks, groups binds by submap, and returns them as `config.Apps`. `main.go` gains a `-H` / `--hyprland` flag that calls `ParseBinds()` and appends the result to the existing `keys` slice before rendering.

**Tech Stack:** Go 1.26, `os/exec`, `encoding/json`, `github.com/kencx/keyb/config`, `flag` package (stdlib), Nix flakes, direnv

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `hyprland/hyprland.go` | Create | `ParseBinds()`, all internal helpers |
| `hyprland/hyprland_test.go` | Create | Unit tests with fixture JSON |
| `main.go` | Modify | Add `-H` flag, call `ParseBinds()`, merge results |
| `flake.nix` | Create | Nix dev shell |
| `.envrc` | Create | direnv activation |

---

## Task 1: Scaffold `hyprland` package with modmask decoder

**Files:**
- Create: `hyprland/hyprland.go`
- Create: `hyprland/hyprland_test.go`

- [ ] **Step 1: Write failing test for `decodeMods`**

Create `hyprland/hyprland_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -run TestDecodeMods -v
```

Expected: compile error — package does not exist yet.

- [ ] **Step 3: Create `hyprland/hyprland.go` with `decodeMods`**

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -run TestDecodeMods -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add hyprland/hyprland.go hyprland/hyprland_test.go
git commit -m "feat(hyprland): add modmask decoder"
```

---

## Task 2: Add `formatKey` and `resolveName` helpers

**Files:**
- Modify: `hyprland/hyprland.go`
- Modify: `hyprland/hyprland_test.go`

- [ ] **Step 1: Write failing tests**

Append to `hyprland/hyprland_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -run "TestFormatKey|TestResolveName" -v
```

Expected: compile error — `formatKey` and `resolveName` undefined.

- [ ] **Step 3: Add `formatKey` and `resolveName` to `hyprland.go`**

Append to `hyprland/hyprland.go`:

```go
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
		return "exec: " + parts[0]
	}
	return dispatcher + ": " + arg
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -run "TestFormatKey|TestResolveName" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add hyprland/hyprland.go hyprland/hyprland_test.go
git commit -m "feat(hyprland): add formatKey and resolveName helpers"
```

---

## Task 3: Implement `ParseBinds` with grouping logic

**Files:**
- Modify: `hyprland/hyprland.go`
- Modify: `hyprland/hyprland_test.go`

- [ ] **Step 1: Write failing test for `ParseBinds`**

Replace the `import` block at the top of `hyprland/hyprland_test.go` with the merged version (Go allows only one import block per file), then append the test function:

```go
// replace existing import block with this merged one:
import (
	"reflect"
	"testing"

	"github.com/kencx/keyb/config"
)
```

Then append the following after the existing test functions:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -run TestParseBindsFromJSON -v
```

Expected: compile error — `parseBindsFromJSON` undefined.

- [ ] **Step 3: Implement `parseBindsFromJSON` and `ParseBinds`**

This is an **additive edit** to `hyprland/hyprland.go` — do NOT overwrite the file. Make two changes:

1. Expand the existing `import "strings"` to the full import block:

```go
import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/kencx/keyb/config"
)
```

2. Append the following new types and functions after the existing helpers:

```go
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
```

- [ ] **Step 4: Run all hyprland tests**

```bash
cd /home/atb/Code/keyb && go test ./hyprland/ -v
```

Expected: all PASS

- [ ] **Step 5: Run full test suite to check nothing broken**

```bash
cd /home/atb/Code/keyb && go test ./...
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add hyprland/hyprland.go hyprland/hyprland_test.go
git commit -m "feat(hyprland): implement ParseBinds with submap grouping"
```

---

## Task 4: Wire `-H` flag into `main.go`

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add the flag and merge logic**

In `main.go`, add the import `"github.com/kencx/keyb/hyprland"` and:

1. Declare `hyprlandMode bool` alongside the other flag vars at the top of `main()`
2. Register the flag pair (after the existing flag declarations):

```go
flag.BoolVar(&hyprlandMode, "H", false, "auto-detect Hyprland keybinds")
flag.BoolVar(&hyprlandMode, "hyprland", false, "auto-detect Hyprland keybinds")
```

3. After the `add` subcommand block (around line 113 — after the `switch args[0]` block that calls `os.Exit(0)`) but before `m := ui.NewModel(keys, cfg)`, add:

```go
if hyprlandMode {
    hyprKeys, err := hyprland.ParseBinds()
    if err != nil {
        log.Fatal(err)
    }
    keys = append(keys, hyprKeys...)
}
```

Also update the help constant to include the new flag:

```go
const help = `usage: keyb [options] <command>

  Options:
    -p, --print       Print to stdout
    -e, --export      Export to file [yaml, json]
    -k, --key         Key bindings at custom path
    -c, --config      Config file at custom path
    -H, --hyprland    Auto-detect Hyprland keybinds
    -v, --version     Version info
    -h, --help        Show help

  Commands:
    a, add            Add keybind to keyb file
`
```

- [ ] **Step 2: Build to verify it compiles**

```bash
cd /home/atb/Code/keyb && go build -o /tmp/keyb-test .
```

Expected: no errors, binary produced.

- [ ] **Step 3: Manual smoke test (requires Hyprland running)**

```bash
/tmp/keyb-test -H -p | head -20
```

Expected: lines of hyprland keybinds printed to stdout including `Super + V`, `Super + Shift + Q` etc.

- [ ] **Step 4: Run full test suite**

```bash
cd /home/atb/Code/keyb && go test ./...
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add main.go
git commit -m "feat: add -H flag for Hyprland keybind auto-detection"
```

---

## Task 5: Add `flake.nix` and `.envrc`

**Files:**
- Create: `flake.nix`
- Create: `.envrc`

- [ ] **Step 1: Create `flake.nix`**

```nix
{
  description = "keyb dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSystem = f: nixpkgs.lib.genAttrs systems
        (system: f nixpkgs.legacyPackages.${system});
    in {
      devShells = forEachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # use go_1_26 if available in nixpkgs, otherwise go (latest stable)
            (go_1_26 or go)
            gopls
            gotools
            golangci-lint
            hyprland
          ];
        };
      });
    };
}
```

> **Note:** `(go_1_26 or go)` is not valid Nix syntax — use `pkgs.go_1_26 or pkgs.go` evaluated at the attribute level. The actual implementation should be:
> ```nix
> (if pkgs ? go_1_26 then pkgs.go_1_26 else pkgs.go)
> ```

Correct `flake.nix`:

```nix
{
  description = "keyb dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSystem = f: nixpkgs.lib.genAttrs systems
        (system: f nixpkgs.legacyPackages.${system});
    in {
      devShells = forEachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = [
            (if pkgs ? go_1_26 then pkgs.go_1_26 else pkgs.go)
            pkgs.gopls
            pkgs.gotools
            pkgs.golangci-lint
            pkgs.hyprland
          ];
        };
      });
    };
}
```

- [ ] **Step 2: Create `.envrc`**

```
use flake
```

- [ ] **Step 3: Verify the flake evaluates**

```bash
cd /home/atb/Code/keyb && nix flake check 2>&1 | head -20
```

If `nix flake check` fails on missing lockfile, run:

```bash
nix flake lock
```

Then re-run check. Expected: no evaluation errors (build errors for missing hyprland pkg are acceptable — the dev shell packages are what matter).

- [ ] **Step 4: Verify the dev shell enters**

```bash
cd /home/atb/Code/keyb && nix develop --command go version
```

Expected: prints Go version.

- [ ] **Step 5: Allow direnv**

```bash
cd /home/atb/Code/keyb && direnv allow
```

Expected: direnv activates the flake dev shell on directory entry.

- [ ] **Step 6: Commit**

```bash
git add flake.nix flake.lock .envrc
git commit -m "chore: add Nix flake and direnv config"
```

---

## Task 6 (Post-implementation, personal config): Add descriptions to Hyprland keybinds

**Files:**
- Modify: `~/.config/hypr/conf/keybinds.conf`

This task is outside the repo. After the feature ships, annotate each bind in the user's personal Hyprland config with a `description` so keyb shows accurate human-readable names instead of the `dispatcher: arg` fallback.

Hyprland supports descriptions via a `# Description:` comment convention — check the [Hyprland wiki on bind descriptions](https://wiki.hyprland.org) for the exact syntax, which may be `bind = ..., desc:My description` as an inline field or a comment-based annotation depending on the Hyprland version.

- [ ] **Step 1: Check Hyprland version for description syntax**

```bash
hyprctl version
```

- [ ] **Step 2: Read current keybinds**

```bash
cat ~/.config/hypr/conf/keybinds.conf
```

- [ ] **Step 3: Add descriptions to each bind**

For each `bind = ...` line, add an accurate, concise description. Examples:

```
bind = $mainMod, return, exec, $terminal # description: open terminal
bind = $mainMod, Q, killactive,          # description: close window
bind = $mainMod, space, exec, ...        # description: open launcher
```

Verify with `keyb -H` that descriptions now appear in the overlay instead of fallback names.

---

## Verification Checklist

- [ ] `go test ./...` passes
- [ ] `keyb -H` shows a "Hyprland" section in the overlay with live keybinds
- [ ] `keyb -H -p` prints hyprland keybinds to stdout
- [ ] `keyb` (without `-H`) behavior is unchanged
- [ ] `nix develop` enters a shell with `go`, `gopls`, `gotools`, `golangci-lint`
- [ ] `direnv allow` in the repo dir activates the dev shell automatically
