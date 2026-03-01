// Package profile provides YAML-based profile loading.
package profile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rlebel12/gomux/internal/gomux"
)

var (
	errMissingName    = errors.New("missing name")
	errMissingWindows = errors.New("missing windows")
)

// yamlFile is the top-level YAML structure.
type yamlFile struct {
	Profiles []yamlProfile `yaml:"profiles"`
}

// yamlProfile maps a single profile entry from YAML.
type yamlProfile struct {
	Name    string       `yaml:"name"`
	Windows []yamlWindow `yaml:"windows"`
}

// yamlWindow maps a window entry from YAML.
type yamlWindow struct {
	Name    string     `yaml:"name"`
	Dir     string     `yaml:"dir"`
	Command string     `yaml:"command"`
	Panes   []yamlPane `yaml:"panes"`
}

// yamlPane maps a pane entry from YAML.
type yamlPane struct {
	Dir     string `yaml:"dir"`
	Command string `yaml:"command"`
	Split   string `yaml:"split"`
}

// Loader implements gomux.ProfileLoader by reading YAML from a file.
type Loader struct {
	path string
}

// NewLoader creates a Loader that reads profiles from the given config file path.
func NewLoader(path string) *Loader {
	return &Loader{path: path}
}

// LoadProfiles reads and parses the YAML config file, returning all defined profiles.
// If the file does not exist, an empty slice is returned without error.
func (l *Loader) LoadProfiles(_ context.Context) ([]gomux.Profile, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []gomux.Profile{}, nil
		}
		return nil, fmt.Errorf("read profiles: %w", err)
	}

	var f yamlFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}

	if err := validateProfiles(f.Profiles); err != nil {
		return nil, err
	}

	return convertProfiles(f.Profiles), nil
}

// validateProfiles checks that all profiles and their windows have required fields.
func validateProfiles(profiles []yamlProfile) error {
	for i, p := range profiles {
		if p.Name == "" {
			return fmt.Errorf("profile %d: %w", i, errMissingName)
		}
		if len(p.Windows) == 0 {
			return fmt.Errorf("profile %q: %w", p.Name, errMissingWindows)
		}
		for j, w := range p.Windows {
			if w.Name == "" {
				return fmt.Errorf("profile %q, window %d: %w", p.Name, j, errMissingName)
			}
		}
	}
	return nil
}

// convertProfiles converts YAML types to domain types, expanding tilde in Dir fields.
func convertProfiles(profiles []yamlProfile) []gomux.Profile {
	result := make([]gomux.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = gomux.Profile{
			Name:    p.Name,
			Windows: convertWindows(p.Windows),
		}
	}
	return result
}

// convertWindows converts YAML window entries to domain Window types.
func convertWindows(windows []yamlWindow) []gomux.Window {
	result := make([]gomux.Window, len(windows))
	for i, w := range windows {
		result[i] = gomux.Window{
			Name:    w.Name,
			Dir:     expandTilde(w.Dir),
			Command: w.Command,
			Panes:   convertPanes(w.Panes),
		}
	}
	return result
}

// convertPanes converts YAML pane entries to domain Pane types.
func convertPanes(panes []yamlPane) []gomux.Pane {
	if len(panes) == 0 {
		return nil
	}
	result := make([]gomux.Pane, len(panes))
	for i, p := range panes {
		result[i] = gomux.Pane{
			Dir:     expandTilde(p.Dir),
			Command: p.Command,
			Split:   p.Split,
		}
	}
	return result
}

// expandTilde replaces a leading "~/" with the user's home directory.
// If the home directory cannot be determined or the path does not start with "~/",
// the path is returned unchanged.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return home + path[1:]
}
