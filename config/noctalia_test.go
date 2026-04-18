package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeNoctaliaColors(t *testing.T) {
	base := Color{
		FilterFg:    "#FFA066",
		BorderColor: "#111111",
	}
	override := Color{
		PromptColor:   "#63dbb5",
		FilterFg:      "#63dbb5",
		FilterBg:      "#162125",
		CursorFg:      "#00382a",
		CursorBg:      "#89d6b9",
		CounterFg:     "#87d1eb",
		PlaceholderFg: "#344a52",
		BorderColor:   "#344a52",
	}
	got := mergeColor(base, override)

	if got.PromptColor != "#63dbb5" {
		t.Errorf("PromptColor: got %q, want %q", got.PromptColor, "#63dbb5")
	}
	if got.FilterFg != "#63dbb5" {
		t.Errorf("FilterFg: got %q, want %q", got.FilterFg, "#63dbb5")
	}
	if got.BorderColor != "#344a52" {
		t.Errorf("BorderColor: got %q, want %q", got.BorderColor, "#344a52")
	}
	if got.CounterBg != "" {
		t.Errorf("CounterBg should be empty (not overridden), got %q", got.CounterBg)
	}
}

func TestLoadNoctaliaColors_Missing(t *testing.T) {
	dir := t.TempDir()
	c, ok, err := loadNoctaliaColors(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false when file is absent")
	}
	if c != (Color{}) {
		t.Errorf("expected zero Color when file is absent, got %+v", c)
	}
}

func TestLoadNoctaliaColors_Present(t *testing.T) {
	dir := t.TempDir()
	yml := []byte(`
color:
  prompt: "#63dbb5"
  filter_fg: "#63dbb5"
  cursor_bg: "#89d6b9"
  border_color: "#344a52"
`)
	if err := os.WriteFile(filepath.Join(dir, "noctalia-colors.yml"), yml, 0644); err != nil {
		t.Fatal(err)
	}

	c, ok, err := loadNoctaliaColors(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ok=true when file is present")
	}
	if c.PromptColor != "#63dbb5" {
		t.Errorf("PromptColor: got %q", c.PromptColor)
	}
	if c.CursorBg != "#89d6b9" {
		t.Errorf("CursorBg: got %q", c.CursorBg)
	}
	if c.FilterBg != "" {
		t.Errorf("FilterBg should be empty (not in file), got %q", c.FilterBg)
	}
}
