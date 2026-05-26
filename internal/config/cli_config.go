package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Launitfile struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Runtime string `yaml:"runtime"`
	Path    string `yaml:"path"`
	Size    string `yaml:"size"`
	Timeout int32  `yaml:"timeout"`
}

type CliConfig interface {
	GetTemplateBytes() ([]byte, error)
	Load() (*Launitfile, error)
}

type YamlCliConfig struct{}

var _ CliConfig = (*YamlCliConfig)(nil)

func NewYamlCliConfig() *YamlCliConfig {
	return &YamlCliConfig{}
}

func (c *YamlCliConfig) GetTemplateBytes() ([]byte, error) {
	defaultConfig := Launitfile{
		App: AppConfig{
			Name:    "example-app",
			Runtime: "go",
			Path:    "./main.go",
			Size:    "small",
			Timeout: 30,
		},
	}

	// Turn the struct into YAML bytes
	yamlBytes, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to generate configuration template: %w", err)
	}

	header := []byte("# Launitfile - Edit this to configure your app\n\n")
	finalPayload := append(header, yamlBytes...)

	return finalPayload, nil
}

func (c *YamlCliConfig) Load() (*Launitfile, error) {
	if _, err := os.Stat("Launitfile"); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("❌ Cannot find Launitfile in current directory. Run \"launit init\" to create one")
		}
		return nil, err
	}

	data, err := os.ReadFile("Launitfile")
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read Launitfile: %w", err)
	}

	cfg := &Launitfile{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf(
			"--------------------------------------------------\n"+
				"❌ LAUNITFILE SYNTAX ERROR DETECTED\n"+
				"--------------------------------------------------\n"+
				"Message: %v\n"+
				"--------------------------------------------------\n"+
				"💡 Tip: Common YAML mistakes include:\n"+
				"   1. Using TABS instead of SPACES for indentation.\n"+
				"   2. Missing a space after a colon (e.g., use 'size: small', NOT 'size:small').\n"+
				"   3. Missing quote marks around special symbols.\n"+
				"--------------------------------------------------",
			err,
		)
	}

	return cfg, nil
}
