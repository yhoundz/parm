package cmdparser

import "testing"

func TestParseRepoRef(t *testing.T) {
	refs := map[string][]string{
		"neovim/neovim":               {"neovim", "neovim", "false"},
		"AvaloniaUI/Avalonia.Samples": {"AvaloniaUI", "Avalonia.Samples", "false"},
		"":                            {"", "", "true"},
		"/":                           {"", "", "true"},
		"godotengine/":                {"", "", "true"},
		";.;:-/godot":                 {"", "", "true"},
	}
	for ref := range refs {
		own, rep, err := ParseRepoRef(ref)
		// TODO: check errors are correct
		if refs[ref][0] != own || refs[ref][1] != rep {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, refs[ref][0], refs[ref][1], err)
			return
		}
	}
}

func TestParseRepoReleaseRef(t *testing.T) {
	refs := map[string][]string{
		"neovim/neovim@v0.11.3":               {"neovim", "neovim", "v0.11.3", "false"},
		"AvaloniaUI/Avalonia.Samples@samples": {"AvaloniaUI", "Avalonia.Samples", "samples", "false"},
		"":                                    {"", "", "", "true"},
		"/@j":                                 {"", "", "", "true"},
		"godotengine/@4.4.1-stable":           {"", "", "", "true"},
		";.;:-/godot@4.4.1-stable":            {"", "", "", "true"},
	}
	for ref := range refs {
		own, rep, tag, err := ParseRepoReleaseRef(ref)
		// TODO: check errors are correct
		if refs[ref][0] != own || refs[ref][1] != rep || refs[ref][2] != tag {
			t.Errorf("error, got '%s'/'%s', wanted %s/%s with err %q", own, rep, refs[ref][0], refs[ref][1], err)
			return
		}
	}
}
