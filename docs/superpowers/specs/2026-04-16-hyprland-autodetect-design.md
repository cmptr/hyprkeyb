# Hyprland Auto-Detection & Nix Dev Environment

**Date:** 2026-04-16  
**Status:** Approved  
**Scope:** Two independent additions to the keyb fork

---

## Summary

Add automatic detection of Hyprland keybinds from the running compositor via `hyprctl binds -j`, surfaced via a `-H` flag. Also add a Nix flake and `.envrc` for a reproducible dev environment.

---

## Feature 1: Hyprland Keybind Detection

### Trigger

A new `-H` / `--hyprland` boolean flag in `main.go`. When passed, keyb queries the running Hyprland compositor and merges detected binds into the display alongside any manually configured keyb file. No change to existing behavior when the flag is absent.

### Architecture

A new self-contained `hyprland/` package with one exported entry point:

```go
func ParseBinds() (config.Apps, error)
```

No other package (except `main.go`) depends on it. The package:

1. Runs `hyprctl binds -j` via `os/exec`
2. Unmarshals JSON into internal `[]hyprBind` structs
3. Decodes each bind's modmask, formats the key combo, resolves a display name
4. Groups binds by submap into `config.Apps`
5. Returns sorted apps ready to merge

### Package layout

```
hyprland/
  hyprland.go       # ParseBinds(), all internal helpers
  hyprland_test.go  # unit tests using fixture JSON
```

### Data pipeline

```
hyprctl binds -j
    │
    ▼
[]hyprBind { modmask, key, dispatcher, arg, description, submap }
    │
    ├─ decodeMods(modmask uint)  → "Super + Shift"
    ├─ formatKey(mods, key)      → "Super + Shift + F"
    └─ resolveName(desc, disp, arg) → description || "dispatcher: arg"
    │
    ▼
group by submap
    ""        → App{Name: "Hyprland"}
    "resize"  → App{Name: "Hyprland: resize"}
    │
    ▼
config.Apps  (sorted: "Hyprland" first, submaps alphabetical)
```

### Modmask decoding

Standard X11 modifier bit values applied in display order:

| Bit | Modifier |
|-----|----------|
| 1   | Shift    |
| 4   | Ctrl     |
| 8   | Alt      |
| 64  | Super    |
| 128 | Mod5     |

Unknown bits are silently ignored (not rendered). Tests must include a fixture with an unrecognized bit to document this behavior explicitly.

Output format: `"Super + Shift + F"` (modifiers in the order above, then the key).

### Name resolution

1. Use `description` field if non-empty (Hyprland opt-in via config annotation)
2. Fallback: `"dispatcher: arg"` — e.g., `"exec: ghostty"`, `"movefocus: l"`, `"killactive"`
   - If `arg` is empty, show only the dispatcher
   - For `exec` binds, show only the first word of arg to keep names short

### Section naming

- Empty submap → section `"Hyprland"`
- Named submap `"resize"` → section `"Hyprland: resize"`

This namespacing makes hyprland-sourced sections visually distinct when mixed with other apps in the keyb YAML.

### Merging

`main.go` appends the result of `hyprland.ParseBinds()` to the `keys` slice returned by `config.Parse()`, then passes the combined list to `ui.NewModel()`. Order: YAML apps first, hyprland apps appended at the end.

### Error handling

- If `hyprctl` is not in `$PATH` or returns a non-zero exit: return a descriptive error surfaced via `log.Fatal` in `main.go`
- If `HYPRLAND_INSTANCE_SIGNATURE` is unset (Hyprland not running): `hyprctl` will fail; the error message will guide the user

### Testing

Unit tests in `hyprland/hyprland_test.go` use a fixture JSON payload (no live `hyprctl` required):

- `TestDecodeMods` — modmask bit combinations
- `TestFormatKey` — full key string assembly
- `TestResolveName` — description present / absent / exec truncation
- `TestParseBinds` — end-to-end from JSON fixture to `config.Apps`

---

## Feature 2: Nix Dev Environment

### `flake.nix`

At repo root. Provides `devShells.default` for `x86_64-linux` and `aarch64-linux`:

- `go` — use `pkgs.go` (latest stable in nixpkgs); if a pinned attribute matching `go.mod`'s `go 1.26.1` exists (e.g., `go_1_26`), prefer it, otherwise `pkgs.go` is sufficient
- `gopls` — language server
- `gotools` — `goimports`, `godoc`
- `golangci-lint` — matches CI linting
- `hyprland` package — provides `hyprctl` for manual local testing

No `flake-utils` dependency; use inline `eachDefaultSystem` pattern to keep inputs minimal.

### `.envrc`

```
use flake
```

Placed at repo root. Activates the dev shell automatically when entering the directory with direnv.

---

## Implementation tasks (ordered)

1. Create `hyprland/hyprland.go` with `ParseBinds()`, `decodeMods()`, `formatKey()`, `resolveName()`, grouping and sorting logic
2. Create `hyprland/hyprland_test.go` with fixture JSON and all unit tests
3. Update `main.go`: add `-H` / `--hyprland` flag, call `hyprland.ParseBinds()`, merge results
4. Create `flake.nix`
5. Create `.envrc`
6. *(Post-implementation / personal config)* Add descriptions to the user's hyprland keybinds config (`~/.config/hypr/conf/keybinds.conf`) — annotate each bind with accurate `description` values so they appear in keyb instead of the `dispatcher: arg` fallback. This is outside the repo and done after the feature ships.

---

## Out of scope

- Parsing hyprland `.conf` files directly (unnecessary given `hyprctl`)
- Config-file option to enable hyprland (flag is sufficient)
- Live reload / IPC watching for bind changes
- Support for `bindm`, `bindr`, `binde` variants (can be added later)
