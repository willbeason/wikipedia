package ingest_wikidata

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/entities"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   "ingest-wikidata workspace_path wikidata_path",
		Short: `extract wikidata into a wikopticon workspace`,
		RunE:  runCmd,
	}

	return cmd
}

var ErrIngestWikidata = errors.New("unable to run wikidata extraction")

const WikidataPrefix = `wikidata`

// WikidataPattern matches wikidata-20240701-all.json.bz2.
var WikidataPattern = regexp.MustCompile(`wikidata-(\d+)-all\.json\.bz2`)

func runCmd(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("%w: not implemented to run outside a workspace", ErrIngestWikidata)
}

func IngestWikidata(cmd *cobra.Command, wikidataCfg *config.IngestWikidata, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: must have exactly one argument but got %+v", ErrIngestWikidata, args)
	}
	wikidataPath := args[0]

	wikidataPathMatches := WikidataPattern.FindStringSubmatch(wikidataPath)
	if len(wikidataPathMatches) < 1 {
		return fmt.Errorf("%w: wikidata path %q does not match known pattern", ErrIngestWikidata, wikidataPath)
	}
	wikidataDate := wikidataPathMatches[1]
	if !strings.HasSuffix(wikidataDate, "1") {
		wikidataDate = wikidataDate[:len(wikidataDate)-1] + "1"
	}

	workspacePath, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	corpusPath := filepath.Join(workspacePath, wikidataDate)
	err = os.MkdirAll(corpusPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIngestWikidata, err)
	}

	// Perform the extraction.
	extractPath := filepath.Join(corpusPath, config.WikidataDir)
	err = flags.CreateOrCheckDirectory(extractPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIngestWikidata, err)
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())
	unparsedObjects, err := source(cancel, wikidataPath)
	if err != nil {
		return err
	}


	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	if len(wikidataCfg.Claims) == 0 {
		return fmt.Errorf("%w: no claims specified", ErrIngestWikidata)
	}

	titleIndexPath := filepath.Join(workspacePath, wikidataDate, wikidataCfg.Index)
	titleIndexFuture := documents.ReadTitleMap(ctx, titleIndexPath, errs)
	titleIndex := <-titleIndexFuture

	parsedEntities := parseObjects(cancel, parallel, wikidataCfg, titleIndex, unparsedObjects)

	outDB, err := badger.Open(badger.DefaultOptions(extractPath))
	if err != nil {
		return fmt.Errorf("opening %q: %w", extractPath, err)
	}
	defer func() {
		closeErr := outDB.Close()
		if err != nil {
			err = closeErr
		}
	}()

	runner := jobs.NewRunner()
	sinkWork := jobs.Reduce(ctx, jobs.WorkBuffer, parsedEntities, db.WriteProto[protos.ID](outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)
	sinkWg.Wait()

	err = db.RunGC(outDB)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	// Close the DB before reading.
	err = outDB.Close()
	if err != nil {
		return fmt.Errorf("closing ingested articles database: %w", err)
	}

	return nil
}

type unparsedObject []byte

func source(cancel context.CancelCauseFunc, wikidata string) (<-chan unparsedObject, error) {
	fWikidata, err := os.Open(wikidata)
	if err != nil {
		return nil, fmt.Errorf("%w: could not open wikidata file: %w", ErrIngestWikidata, err)
	}

	unparsedObjects := make(chan unparsedObject, 10000)

	go func() {
		defer func() {
			err = fWikidata.Close()
			if err != nil {
				cancel(err)
			}
		}()

		var rWikidata *bufio.Reader
		compressed := filepath.Ext(wikidata) == ".bz2"
		if compressed {
			rWikidata = bufio.NewReader(bzip2.NewReader(fWikidata))
		} else {
			rWikidata = bufio.NewReader(fWikidata)
		}

		err = extractFile(rWikidata, unparsedObjects)
		if err != nil {
			cancel(err)
		}

		close(unparsedObjects)
	}()

	return unparsedObjects, nil
}

func extractFile(r *bufio.Reader, unparsedObjects chan<- unparsedObject) error {
	_, err := r.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("%w: reading first line: %w", ErrIngestWikidata, err)
	}
	// Skip first line.

	var line []byte
	for line, err = r.ReadBytes('\n'); err == nil; line, err = r.ReadBytes('\n') {
		unparsedObjects <- []byte(strings.TrimRight(string(line), ",]"))
	}

	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: reading line: %w", ErrIngestWikidata, err)
	}

	// Skip last line since it just closes the massive list of entities.

	return nil
}

func parseObjects(
	cancel context.CancelCauseFunc,
	parallel int,
	cfg *config.IngestWikidata,
	index map[string]uint32,
	unparsedObjects <-chan unparsedObject,
) <-chan protos.ID {
	parsedObjects := make(chan protos.ID, 10000)

	wg := sync.WaitGroup{}
	for range parallel {
		wg.Add(1)
		go func() {
			err := parseObjectsWorker(cfg, index, unparsedObjects, parsedObjects)
			if err != nil {
				cancel(err)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(parsedObjects)
	}()

	return parsedObjects
}

var count atomic.Int64

func parseObjectsWorker(
	cfg *config.IngestWikidata,
	index map[string]uint32,
	unparsedObjects <-chan unparsedObject,
	parsedEntities chan<- protos.ID,
) error {
	allowedInstances := make(map[string]bool)
	for _, instanceOf := range cfg.InstanceOf {
		allowedInstances[instanceOf] = true
	}
	requireWikipediaArticle := make(map[string]bool)
	for _, r := range cfg.RequireWikipedia {
		requireWikipediaArticle[r] = true
	}

	for unparsedObj := range unparsedObjects {
		if len(unparsedObj) == 0 {
			// Empty line or end of input.
			continue
		}
		parsedEntity, err := ParseObject(allowedInstances, requireWikipediaArticle, cfg.Claims, index, unparsedObj)
		if errors.Is(err, ErrNotObject) {
			continue
		}
		if err != nil {
			return err
		}
		if parsedEntity == nil {
			continue
		}

		parsedEntities <- parsedEntity

		newCount := count.Add(1)
		if newCount%1000 == 0 {
			fmt.Println("Count", newCount, parsedEntity.Id, parsedEntity.WikidataId, parsedEntity.Sitelinks["enwiki"].Title)
		}
	}

	return nil
}

var ErrNotObject = errors.New("not an object")

const (
	InstanceOf = "P31"
)

func ParseObject(
	allowedInstances map[string]bool,
	requireWikipediaArticle map[string]bool,
	keepClaims []string,
	index map[string]uint32,
	unparsedObj []byte,
) (*entities.Entity, error) {
	rawEntity := &Entity{}
	err := json.Unmarshal(unparsedObj, &rawEntity)
	if err != nil {
		return nil, fmt.Errorf("%w: unmarshalling json %q: %w", ErrIngestWikidata, string(unparsedObj), err)
	}
	wikidataId := rawEntity.Id

	parsedEntity := &entities.Entity{
		WikidataId: wikidataId,
		Claims:     make(map[string]*entities.Claims),
	}

	instanceOf, hasInstanceOf := rawEntity.Claims[InstanceOf]
	if !hasInstanceOf {
		// Not an instance of anything, so skip it.
		return nil, ErrNotObject
	}

	wantInstance := false
	requireWikipedia := false
	for _, claim := range instanceOf {
		if claim.Mainsnak.DataValue.Value == nil {
			// This instanceOf claim is empty.
			continue
		}

		value, isValue := claim.Mainsnak.DataValue.Value.(map[string]interface{})
		if !isValue {
			fmt.Printf("Strange claim for %q: %+v\n", wikidataId, claim)
			panic("Strange claim")
		}

		valueStr, isStr := value["id"].(string)
		if !isStr {
			fmt.Printf("Strange claim for %q: %+v\n", wikidataId, claim)
			panic("Strange claim")
		}

		if allowedInstances[valueStr] {
			wantInstance = true
		}
		if requireWikipediaArticle[valueStr] {
			requireWikipedia = true
		}
	}
	if !wantInstance {
		// Not an instance of anything we care about, so skip it.
		return nil, ErrNotObject
	}

	var title string
	if enwikiSiteLink, ok := rawEntity.SiteLinks["enwiki"]; ok {
		title = enwikiSiteLink.Title
		parsedEntity.Sitelinks = map[string]*entities.SiteLink{
			"enwiki": {
				Site:  enwikiSiteLink.Site,
				Title: title,
				Url:   enwikiSiteLink.Url,
			},
		}
	} else if requireWikipedia {
		// Not on English Wikipedia
		return nil, ErrNotObject
	}
	if requireWikipedia && title == "" {
		// No title listed on English Wikipedia
		return nil, ErrNotObject
	}

	relevantClaim := false
	for _, claimID := range keepClaims {
		if claim, ok := rawEntity.Claims[claimID]; ok {
			entityClaims := &entities.Claims{}

			for _, c := range claim {
				if c.Mainsnak.DataValue.Value == nil {
					continue
				}
				relevantClaim = true

				parseClaims(c, wikidataId, claimID, entityClaims)
			}

			parsedEntity.Claims[claimID] = entityClaims
		}
	}
	if !relevantClaim {
		// No relevant claims.
		return nil, ErrNotObject
	}

	id := index[title]
	if id == 0 {
		// Probably an article in a namespace we don't care about, like Categories.
		return nil, ErrNotObject
	}
	parsedEntity.Id = id

	return parsedEntity, nil
}

func parseClaims(c Claim, wikidataId string, claimID string, entityClaims *entities.Claims) {
	value, isValue := c.Mainsnak.DataValue.Value.(map[string]interface{})
	if !isValue {
		fmt.Printf("Strange claim for %q: %+v\n", wikidataId, c)
		panic("Strange claim")
	}

	var valueStr string

	switch claimID {
	case "P569", "P570":
		var isStr bool
		valueStr, isStr = value["time"].(string)
		if !isStr {
			fmt.Printf("Strange claim for %q: %+v\n", wikidataId, c)
			panic("Strange claim")
		}
	default:
		var isStr bool
		valueStr, isStr = value["id"].(string)
		if !isStr {
			fmt.Printf("Strange claim for %q: %+v\n", wikidataId, c)
			panic("Strange claim")
		}
	}

	entityClaims.Claim = append(entityClaims.Claim, &entities.Claim{
		Property: c.Mainsnak.Property,
		Value:    valueStr,
		Rank:     c.Rank,
	})
}

type Entity struct {
	Id        string              `json:"id"`
	SiteLinks map[string]SiteLink `json:"sitelinks"`
	Claims    map[string][]Claim  `json:"claims"`
}

type SiteLink struct {
	Site  string `json:"site"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Claim struct {
	Mainsnak Mainsnak `json:"mainsnak"`
	Rank     string   `json:"rank"`
}

type Mainsnak struct {
	Property  string    `json:"property"`
	DataValue DataValue `json:"datavalue"`
}

type DataValue struct {
	Value any `json:"value"`
}
