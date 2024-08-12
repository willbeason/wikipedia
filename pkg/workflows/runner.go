package workflows

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/clean"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/links"
	title_index "github.com/willbeason/wikipedia/pkg/title-index"
)

var (
	ErrWorkflowNotExist = errors.New("workflow does not exist")
	ErrWorkflow         = errors.New("unable to run workflow")
	ErrJob              = errors.New("unable to run job")
)

type Runner struct {
	Config *config.Config
}

func (r *Runner) RunCorpusWorkflow(cmd *cobra.Command, corpusName, workflowName string) error {
	if corpusName == "" {
		return fmt.Errorf("%w, unable to run workflow %q: corpusName must be non-empty",
			ErrWorkflow, workflowName)
	}

	workflow, exists := r.Config.Workflows[workflowName]
	if !exists {
		return fmt.Errorf("%w: %q",
			ErrWorkflowNotExist, workflowName)
	}

	for _, jobName := range workflow {
		err := r.RunCorpusJob(cmd, corpusName, jobName)
		if err != nil {
			return fmt.Errorf("running job %q in workflow %q: %w",
				jobName, workflowName, err)
		}
	}

	return nil
}

func (r *Runner) RunCorpusJob(cmd *cobra.Command, corpusName, jobName string) error {
	if corpusName == "" {
		return fmt.Errorf("%w, unable to run job %q: corpusName must be non-empty",
			ErrWorkflow, jobName)
	}

	job, exists := r.Config.Jobs[jobName]
	if !exists {
		return fmt.Errorf("%w: job %q does not exist",
			config.ErrLoad, jobName)
	}

	switch job.SubCommand {
	case "":
		return fmt.Errorf("%w: job %q has no subCommand",
			config.ErrLoad, jobName)
	case "clean":
		cleanCfg, err := config.UnmarshallJob[config.Clean](job)
		if err != nil {
			return err
		}

		return clean.Clean(cmd, corpusName, cleanCfg.In, cleanCfg.Out)
	case "links":
		linksCfg, err := config.UnmarshallJob[config.Links](job)
		if err != nil {
			return err
		}

		return links.Links(cmd, corpusName, linksCfg.In, linksCfg.Index, linksCfg.Out)
	case "title-index":
		titleCfg, err := config.UnmarshallJob[config.TitleIndex](job)
		if err != nil {
			return err
		}

		return title_index.TitleIndex(cmd, corpusName, titleCfg.In, titleCfg.Out)
	default:
		return fmt.Errorf("%w: job %q has unknown subCommand %q",
			config.ErrLoad, jobName, job.SubCommand)
	}
}
