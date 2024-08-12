package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	Filename           = "config.yaml"
	ArticlesDir        = "articles"
	PostIngestWorkflow = "post-ingest"
)

// Config is the configuration for a particular workspace for analyzing one or more Wikipedia corpora.
type Config struct {
	// Ingest configures ingesting new Wikipedia dumps into this workspace.
	Ingest Ingest `yaml:"ingest"`

	// Workflows is a map from workflow name to a list of Jobs to run sequentially.
	Workflows map[string][]string

	// Jobs is a map of directory names representing jobs to the configuration
	// for those jobs. Should be retrieved individually with GetJob.
	//
	// The job name (the key in this map) is the argument which should be passed
	// after the configuration argument.
	Jobs map[string]*Job `yaml:"jobs"`
}

// ErrLoad indicates the config for a job could not be loaded successfully.
var ErrLoad = errors.New("unable to load config")

// Load reads a generic Config object from a file.
func Load(workspacePath string) (*Config, error) {
	config := &Config{}

	configPath := filepath.Join(workspacePath, Filename)

	in, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to read config file", ErrLoad)
	}
	err = yaml.Unmarshal(in, config)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to parse config file", ErrLoad)
	}

	return config, nil
}

// Job represents a generic task to run with all specified configuration.
type Job struct {
	// SubCommand is the subCommand of wikopticon to run. Should correspond
	// one-to-one with a configuration type to unmarshall to. So the subCommand
	// "extract" should map to the "config.Extract" type.
	SubCommand string `yaml:"subCommand"`

	// Settings are the configuration used for a job.
	// Should be unmarshalled into a real config object with unmarshall.
	Settings map[string]interface{} `yaml:"settings"`
}

// UnmarshallJob attempts to extract the Job's config into T.
func UnmarshallJob[T any](j *Job) (*T, error) {
	in, err := yaml.Marshal(j.Settings)
	if err != nil {
		// Should be impossible.
		return nil, fmt.Errorf("marshalling: %w", err)
	}

	var jobConfig T
	err = yaml.Unmarshal(in, &jobConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling to subCommand config %T: %w", jobConfig, err)
	}

	return &jobConfig, nil
}
