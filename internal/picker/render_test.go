package picker

import (
	"os"
	"strings"
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/stretchr/testify/assert"
)

func renderableTestModel() *Model {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "git", Apt: "git"},
				{Name: "only-brew", Brew: "x"},
			}},
			"07-gpu": {When: "has_gpu", Tools: []config.Tool{
				{Name: "cuda", Apt: "cuda"},
			}},
		},
	}
	return New(cfg, linuxAptInfo())
}

func TestRender_ContainsProfileNames(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "00-base")
	assert.Contains(t, out, "07-gpu")
}

func TestRender_ExpandedShowsTools(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	// Expand 00-base.
	for _, n := range m.Nodes {
		if n.Kind == KindProfile && n.Name == "00-base" {
			n.Expanded = true
		}
	}
	out := Render(m, 80, 24)
	assert.Contains(t, out, "git")
	assert.Contains(t, out, "only-brew")
}

func TestRender_CollapsedHidesTools(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.NotContains(t, out, "  git") // tool indent wouldn't appear when collapsed
}

func TestRender_CheckboxesForStates(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "[x]", "checked profile")
	assert.Contains(t, out, "[·]", "disabled profile")
}

func TestRender_DisabledProfileShowsMarker(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "no GPU detected", "disabled profile displays the reason")
}

func TestRender_NoColorByDefault(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.NotContains(t, out, "\x1b[", "NO_COLOR=1 strips ANSI")
}

func TestRender_ColorWhenEnabled(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "\x1b[", "ANSI codes present when color is enabled")
}

func TestRender_StatusBarIncludesKeys(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "Space")
	assert.Contains(t, out, "Enter")
	assert.Contains(t, out, "q")
}

func TestRender_NarrowWidthDropsMethodColumn(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	for _, n := range m.Nodes {
		if n.Kind == KindProfile && n.Name == "00-base" {
			n.Expanded = true
		}
	}
	wide := Render(m, 80, 24)
	narrow := Render(m, 40, 24)
	// Wide shows "apt" next to git; narrow drops it.
	assert.Contains(t, wide, "apt")
	assert.NotContains(t, narrow, "apt")
	// Both still show the tool name.
	assert.Contains(t, narrow, "git")
}

func TestRender_StatusMessageAppears(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	m.StatusMsg = "Nothing selected — press Space to select tools"
	out := Render(m, 80, 24)
	assert.Contains(t, out, "Nothing selected")
}

func TestRender_HelpOverlay(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	m.HelpOpen = true
	out := Render(m, 80, 24)
	// Help overlay should include more detailed key documentation.
	assert.Contains(t, strings.ToLower(out), "help")
	assert.Contains(t, out, "↑/↓")
	assert.Contains(t, out, "Space")
}

func TestRenderConfirm_ListsSelectedAndDeps(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "curl", Apt: "curl"},
			}},
			"04-ai": {Tools: []config.Tool{
				{Name: "claude-code", DependsOn: "curl", Apt: "claude-code"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	// Uncheck curl so we can verify it comes back via dep resolution.
	for _, n := range m.Nodes {
		if n.Name == "curl" {
			n.Check = Unchecked
		}
	}
	for _, n := range m.Nodes {
		if n.Kind == KindProfile {
			n.Check = profileAggregate(m, n)
		}
	}
	m.Phase = PhaseConfirm
	out := Render(m, 80, 24)

	assert.Contains(t, out, "claude-code")
	assert.Contains(t, out, "curl")
	assert.Contains(t, out, "Required dependencies")
	assert.Contains(t, out, "[Y/Enter]")
	assert.Contains(t, out, "[E]")
	assert.Contains(t, out, "[N/Esc]")
}

func TestRenderConfirm_NoDepsSection(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "git", Apt: "git"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	m.Phase = PhaseConfirm
	out := Render(m, 80, 24)

	assert.Contains(t, out, "git")
	assert.NotContains(t, out, "Required dependencies", "no deps to add → omit section")
}

func TestRenderConfirm_AlreadyInstalledAnnotation(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// "sh" is guaranteed to be in PATH on every Linux/macOS host and CI machine.
	// executor.IsInstalled looks up PATH for unknown tool names, so naming a tool
	// "sh" exercises the "already installed" annotation path reliably.
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "sh", Apt: "sh"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	m.Phase = PhaseConfirm
	out := Render(m, 80, 24)

	assert.Contains(t, out, "already installed, will skip",
		"tool found in PATH should show the already-installed annotation")
}
