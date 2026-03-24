package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  int                `yaml:"version"`
	Settings Settings           `yaml:"settings"`
	Profiles map[string]Profile `yaml:"profiles"`
}

type Settings struct {
	LogLevel string `yaml:"log_level"`
	DryRun   bool   `yaml:"dry_run"`
	Force    bool   `yaml:"-"` // CLI-only, not in YAML
}

type Profile struct {
	When  string `yaml:"when,omitempty"`
	Tools []Tool `yaml:"tools"`
}

type Tool struct {
	Name      string   `yaml:"name"`
	DependsOn string   `yaml:"depends_on,omitempty"`
	Apt       string   `yaml:"apt,omitempty"`
	Dnf       string   `yaml:"dnf,omitempty"`
	Pacman    string   `yaml:"pacman,omitempty"`
	Brew      string   `yaml:"brew,omitempty"`
	Choco     string   `yaml:"choco,omitempty"`
	Winget    string   `yaml:"winget,omitempty"`
	Pip       []string `yaml:"pip,omitempty"`
	Npm       string   `yaml:"npm,omitempty"`
	Shell     Shell    `yaml:"shell,omitempty"`
}

type Shell struct {
	Unix    string `yaml:"unix,omitempty"`
	Linux   string `yaml:"linux,omitempty"`
	Darwin  string `yaml:"darwin,omitempty"`
	Windows string `yaml:"windows,omitempty"`
}

// UnmarshalYAML supports both string and struct forms:
//
//	shell: "echo hello"        → Shell{Unix: "echo hello"}
//	shell:
//	  unix: "echo hello"       → Shell{Unix: "echo hello"}
func (s *Shell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err == nil {
		s.Unix = str
		return nil
	}
	type raw Shell
	return unmarshal((*raw)(s))
}

func LoadFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &cfg, nil
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return LoadFromBytes(data)
}
