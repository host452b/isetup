package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() *Config {
	return &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"base": {Tools: []Tool{{Name: "git", Apt: "git"}}},
		},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	errs, warns := Validate(validConfig())
	assert.Empty(t, errs)
	assert.Empty(t, warns)
}

func TestValidate_MissingVersion(t *testing.T) {
	cfg := validConfig()
	cfg.Version = 0
	errs, _ := Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "version")
}

func TestValidate_UnsupportedVersion(t *testing.T) {
	cfg := validConfig()
	cfg.Version = 99
	errs, _ := Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "version")
}

func TestValidate_DuplicateToolNames(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"a": {Tools: []Tool{{Name: "git", Apt: "git"}}},
			"b": {Tools: []Tool{{Name: "git", Brew: "git"}}},
		},
	}
	errs, _ := Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "duplicate tool name")
}

func TestValidate_MissingToolName(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"base": {Tools: []Tool{{Apt: "git"}}},
		},
	}
	errs, _ := Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "missing name")
}

func TestValidate_CircularDependsOn(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"base": {Tools: []Tool{
				{Name: "a", DependsOn: "b", Apt: "a"},
				{Name: "b", DependsOn: "a", Apt: "b"},
			}},
		},
	}
	errs, _ := Validate(cfg)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "circular")
}

func TestValidate_UnknownWhenCondition(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"test": {When: "has_magic", Tools: []Tool{{Name: "x", Apt: "x"}}},
		},
	}
	errs, _ := Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "unknown when")
}

func TestValidate_ValidWhenCondition(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"gpu": {When: "has_gpu", Tools: []Tool{{Name: "cuda", Apt: "cuda"}}},
		},
	}
	errs, _ := Validate(cfg)
	assert.Empty(t, errs)
}

func TestValidate_DependsOnNonexistent_Warns(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"base": {Tools: []Tool{{Name: "node", DependsOn: "nvm", Apt: "node"}}},
		},
	}
	errs, warns := Validate(cfg)
	assert.Empty(t, errs)
	require.Len(t, warns, 1)
	assert.Contains(t, warns[0], "depends_on")
}

func TestValidate_EmptyToolsList_Warns(t *testing.T) {
	cfg := &Config{
		Version:  1,
		Settings: Settings{LogLevel: "info"},
		Profiles: map[string]Profile{
			"empty": {Tools: []Tool{}},
		},
	}
	errs, warns := Validate(cfg)
	assert.Empty(t, errs)
	require.Len(t, warns, 1)
	assert.Contains(t, warns[0], "empty")
}
