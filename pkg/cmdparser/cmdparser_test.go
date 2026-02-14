package cmdparser

import (
	"strconv"
	"testing"
)

func TestParseRepoRef(t *testing.T) {
	refs := map[string][]string{
		"neovim/neovim":               {"neovim", "neovim", "false"},
		"AvaloniaUI/Avalonia.Samples": {"AvaloniaUI", "Avalonia.Samples", "false"},
		// Shorthand syntax tests
		"neovim":     {"neovim", "neovim", "false"},
		"rust-lang":  {"rust-lang", "rust-lang", "false"},
		"":           {"", "", "true"},
		"/":          {"", "", "true"},
		"godotengine/":               {"", "", "true"},
		";.;:-/godot":                 {"", "", "true"},
	}
	for ref, val := range refs {
		own, rep, err := ParseRepoRef(ref)
		expErr, _ := strconv.ParseBool(val[2])
		actErr := expErr != (err != nil)
		if val[0] != own || val[1] != rep || actErr {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, val[0], val[1], err)
			return
		}
	}
}

func TestParseRepoReleaseRef(t *testing.T) {
	refs := map[string][]string{
		"neovim/neovim@v0.11.3":               {"neovim", "neovim", "v0.11.3", "false"},
		"AvaloniaUI/Avalonia.Samples@samples": {"AvaloniaUI", "Avalonia.Samples", "samples", "false"},
		// Shorthand syntax with tag tests
		"neovim@v0.11.3":      {"neovim", "neovim", "v0.11.3", "false"},
		"rust-lang@1.75.0":    {"rust-lang", "rust-lang", "1.75.0", "false"},
		"":                                    {"", "", "", "true"},
		"/@j":                                 {"", "", "", "true"},
		"godotengine/@4.4.1-stable":           {"", "", "", "true"},
		";.;:-/godot@4.4.1-stable":            {"", "", "", "true"},
	}
	for ref, val := range refs {
		own, rep, tag, err := ParseRepoReleaseRef(ref)
		expErr, _ := strconv.ParseBool(val[3])
		actErr := expErr != (err != nil)
		if val[0] != own || val[1] != rep || val[2] != tag || actErr {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, val[0], val[1], err)
			return
		}
	}
}

func TestParseGithubUrlPattern(t *testing.T) {
	refs := map[string][]string{
		// https
		"https://github.com/neovim/neovim.git":               {"neovim", "neovim", "false"},
		"https://github.com/AvaloniaUI/Avalonia.Samples.git": {"AvaloniaUI", "Avalonia.Samples", "false"},
		"https://github.com/.git":                            {"", "", "true"},
		"https://github.com//.git":                           {"", "", "true"},
		"https://github.com/godotengine/.git":                {"", "", "true"},
		"https://github.com/;.;:-/godot.git":                 {"", "", "true"},

		// ssh
		"git@github.com:neovim/neovim.git":               {"neovim", "neovim", "false"},
		"git@github.com:AvaloniaUI/Avalonia.Samples.git": {"AvaloniaUI", "Avalonia.Samples", "false"},
		"git@github.com:.git":                            {"", "", "true"},
		"git@github.com:/.git":                           {"", "", "true"},
		"git@github.com:godotengine/.git":                {"", "", "true"},
		"git@github.com:;.;:-/godot.git":                 {"", "", "true"},
	}
	for ref, val := range refs {
		own, rep, err := ParseGithubUrlPattern(ref)
		expErr, _ := strconv.ParseBool(val[2])
		actErr := expErr != (err != nil)
		if val[0] != own || val[1] != rep || actErr {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, val[0], val[1], err)
			return
		}
	}
}

func TestParseGithubUrlPatternWithRelease(t *testing.T) {
	refs := map[string][]string{
		// https
		"https://github.com/neovim/neovim.git@v0.11.3":              {"neovim", "neovim", "v0.11.3", "false"},
		"https://github.com/AvaloniaUI/Avalonia.Samples.git@v0.1.1": {"AvaloniaUI", "Avalonia.Samples", "v0.1.1", "false"},
		"https://github.com/.git@":                                  {"", "", "", "true"},
		"https://github.com//.git@i@tag":                            {"", "", "", "true"},
		"https://github.com/godotengine/.git@tt0.v1":                {"", "", "", "true"},
		"https://github.com/;.;:-/godot.git@irn":                    {"", "", "", "true"},

		// ssh
		"git@github.com:neovim/neovim.git@v0.11.3":              {"neovim", "neovim", "v0.11.3", "false"},
		"git@github.com:AvaloniaUI/Avalonia.Samples.git@v0.1.1": {"AvaloniaUI", "Avalonia.Samples", "v0.1.1", "false"},
		"git@github.com:.git@":                                  {"", "", "", "true"},
		"git@github.com:/.git@i@tag":                            {"", "", "", "true"},
		"git@github.com:godotengine/.git@tt0.v1":                {"", "", "", "true"},
		"git@github.com:;.;:-/godot.git@irn":                    {"", "", "", "true"},
	}
	for ref, val := range refs {
		own, rep, tag, err := ParseGithubUrlPatternWithRelease(ref)
		expErr, _ := strconv.ParseBool(val[3])
		actErr := expErr != (err != nil)
		if val[0] != own || val[1] != rep || val[2] != tag || actErr {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, val[0], val[1], err)
			return
		}
	}
}

func TestStringToString(t *testing.T) {
	refs := map[string][]string{
		// https
		"key=value":     {"key", "value", "false"},
		"key=value=key": {"key", "value=key", "false"},
		"hello world":   {"", "", "true"},
		"======":        {"", "=====", "false"},
		"hello = world": {"hello ", " world", "false"},
		"":              {"", "", "true"},
		"six=seven":     {"six", "seven", "false"},
	}
	for ref, val := range refs {
		s1, s2, err := StringToString(ref)
		act1, act2 := val[0], val[1]
		expErr, _ := strconv.ParseBool(val[2])
		actErr := expErr != (err != nil)
		if actErr || s1 != act1 || s2 != act2 {
			t.Errorf("error: got %s and %s, wanted %s and %s. returned with err: %q", s1, s2, act1, act2, err)
		}
	}
}

func TestBuildGitLink(t *testing.T) {
	owner := "foo"
	repo := "bar"
	h, s := BuildGitLink(owner, repo)
	if h != "https://github.com/foo/bar.git" {
		t.Fatalf("unexpected https link: %s", h)
	}
	if s != "git@github.com:foo/bar.git" {
		t.Fatalf("unexpected ssh link: %s", s)
	}
}
