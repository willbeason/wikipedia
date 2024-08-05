package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// WorkPath is an exact path to the working directory to store intermediate
	// and final outputs. All child paths are assumed to be subdirectories of
	// this path.
	WorkPath string `yaml:"workPath"`

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
func Load(configPath string) (*Config, error) {
	config := &Config{}

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

// GetJob extracts the configuration for a subCommand.
func (c *Config) GetJob(name string) (any, error) {
	job, exists := c.Jobs[name]
	if !exists {
		return nil, fmt.Errorf("%w: job %q does not exist", ErrLoad, name)
	}

	var config JobConfig
	switch job.SubCommand {
	case "":
		return nil, fmt.Errorf("%w: job %q has no subCommand",
			ErrLoad, name)
	case "extract":
		config = &Extract{}
	case "clean":
		config = &Clean{}
	default:
		return nil, fmt.Errorf("%w: job %q has unknown subCommand %q",
			ErrLoad, name, job.SubCommand)
	}

	err := job.unmarshall(config)
	if err != nil {
		return nil, fmt.Errorf("%w: unmarshalling job config %q to %T: %w",
			ErrLoad, name, config, err)
	}

	if config.GetWorkPath() == "" {
		// Inherit from parent if not set.
		config.SetWorkPath(c.WorkPath)
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

// unmarshall attempts to extract the Job's config into out.
func (j *Job) unmarshall(out interface{}) error {
	in, err := yaml.Marshal(j.Settings)
	if err != nil {
		// Should be impossible.
		return fmt.Errorf("marshalling: %w", err)
	}

	err = yaml.Unmarshal(in, out)
	if err != nil {
		return fmt.Errorf("unmarshalling to subCommand config %T: %w", out, err)
	}

	return nil
}
