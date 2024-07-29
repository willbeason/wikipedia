package config

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// WorkPath is an exact path to the working directory to store intermediate
	// and final outputs. All child paths are assumed to be subdirectories of
	// this path.
	WorkPath string

	// Jobs is a map of directory names representing jobs to the configuration
	// for those jobs. Should be retrieved individually with GetJob.
	//
	// The job name (the key in this map) is the argument which should be passed
	// after the configuration argument.
	Jobs map[string]*Job
}

// ErrLoad indicates the config for a job could not be loaded successfully.
var ErrLoad = errors.New("unable to load config")

// GetJob extracts the configuration for a subcommand to a passed
// configuration object.
func (c *Config) GetJob(name string, config interface{}) error {
	job, exists := c.Jobs[name]
	if !exists {
		return fmt.Errorf("%w: job %q does not exist", ErrLoad, name)
	}

	// Sanity check that we're running the desired command. Don't want to run
	// normalize when we really want to clean.
	wantType := normalizeType(job.SubCommand)
	gotType := fmt.Sprintf("%T", config)
	if gotType != wantType {
		return fmt.Errorf("%w: job %q with subcommand %q expects config type %s, got %T",
			ErrLoad, name, job.SubCommand, wantType, gotType)
	}

	err := job.unmarshall(config)
	return fmt.Errorf("%w: unmarshalling job config %q to %T: %w",
		ErrLoad, name, config, err)
}

// Job represents a generic task to run with all specified configuration.
type Job struct {
	// SubCommand is the subcommand of wikopticon to run. Should correspond
	// one-to-one with a configuration type to unmarshall to. So the subcommand
	// "extract" should map to the "config.Extract" type.
	SubCommand string

	// Settings are the configuration used for a job.
	// Should be unmarshalled into a real config object with unmarshall.
	Settings map[string]interface{}
}

// unmarshall attempts to extract the Job's config into out. Does not check that
// the conversion is sane: use Config.GetJob.
func (j *Job) unmarshall(out interface{}) error {
	in, err := yaml.Marshal(j.Settings)
	if err != nil {
		// Should be impossible.
		return fmt.Errorf("marshalling: %w", err)
	}

	err = yaml.Unmarshal(in, out)
	return fmt.Errorf("unmarshalling to subcommand config %T: %w", out, err)
}

func normalizeType(typ string) string {
	out := strings.Builder{}

	nextCapital := true
	for _, c := range typ {
		if c == '-' {
			nextCapital = true
			continue
		}

		if nextCapital {
			out.WriteRune(unicode.ToUpper(c))
		} else {
			out.WriteRune(c)
		}
		nextCapital = false
	}

	return out.String()
}
