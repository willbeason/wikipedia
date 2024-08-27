package workflows

import (
	"errors"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/analysis/pagerank"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/analysis/gender"
	"github.com/willbeason/wikipedia/pkg/clean"
	"github.com/willbeason/wikipedia/pkg/config"
	ingest_wikidata "github.com/willbeason/wikipedia/pkg/ingest-wikidata"
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

func (r *Runner) RunWorkflow(cmd *cobra.Command, workflowName string, corpusNames ...string) error {
	workflow, exists := r.Config.Workflows[workflowName]
	if !exists {
		return fmt.Errorf("%w: %q",
			ErrWorkflowNotExist, workflowName)
	}

	for _, jobName := range workflow {
		err := r.RunJob(cmd, jobName, corpusNames...)
		if err != nil {
			return fmt.Errorf("running job %q in workflow %q: %w",
				jobName, workflowName, err)
		}
	}

	return nil
}

func (r *Runner) RunJob(cmd *cobra.Command, jobName string, args ...string) error {
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

		return clean.Clean(cmd, cleanCfg, args...)

	case "gender-frequency":
		cfg, err := config.UnmarshallJob[config.GenderFrequency](job)
		if err != nil {
			return err
		}

		return gender.Frequency(cmd, cfg, args...)
	case "gender-comparison":
		cfg, err := config.UnmarshallJob[config.GenderComparison](job)
		if err != nil {
			return err
		}

		return gender.Comparison(cmd, cfg, args...)

	case "gender-index":
		cfg, err := config.UnmarshallJob[config.GenderIndex](job)
		if err != nil {
			return err
		}

		return gender.Index(cmd, cfg, args...)
	case "links":
		linksCfg, err := config.UnmarshallJob[config.Links](job)
		if err != nil {
			return err
		}

		return links.Links(cmd, linksCfg, args...)
	case "title-index":
		titleCfg, err := config.UnmarshallJob[config.TitleIndex](job)
		if err != nil {
			return err
		}

		return title_index.TitleIndex(cmd, titleCfg, args...)
	case "ingest-wikidata":
		wikidataCfg, err := config.UnmarshallJob[config.IngestWikidata](job)
		if err != nil {
			return err
		}

		return ingest_wikidata.IngestWikidata(cmd, wikidataCfg, args...)
	case "pagerank":
		wikidataCfg, err := config.UnmarshallJob[config.PageRank](job)
		if err != nil {
			return err
		}

		return pagerank.RunPageRank(cmd, wikidataCfg, args...)
	default:
		return fmt.Errorf("%w: job %q has unknown subCommand %q",
			config.ErrLoad, jobName, job.SubCommand)
	}
}
