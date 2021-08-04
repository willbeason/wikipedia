package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/jobs"

	"github.com/willbeason/extract-wikipedia/pkg/flags"

	"github.com/spf13/cobra"
)

const (
	shardSize = 1000
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			nParallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			repo := args[0]
			index := args[1]
			outDir := args[2]

			fRepo, err := os.Open(repo)
			if err != nil {
				return err
			}
			defer func() {
				err = fRepo.Close()
			}()

			fIndex, err := os.Open(index)
			if err != nil {
				return err
			}
			defer func() {
				_ = fIndex.Close()
			}()

			rIndex := bufio.NewReader(fIndex)

			errs, errsWg := jobs.Errors()

			jobFiles := make(chan job)

			jobSync := sync.WaitGroup{}
			jobSync.Add(1)
			go func() {
				extractFile(rIndex, fRepo, jobFiles, errs)

				close(jobFiles)
				jobSync.Done()
			}()

			wg := sync.WaitGroup{}
			for w := 0; w < nParallel; w++ {
				wg.Add(1)
				go func() {
					worker(outDir, jobFiles, errs)
					wg.Done()
				}()
			}

			jobSync.Wait()
			wg.Wait()
			close(errs)
			errsWg.Wait()

			return fIndex.Close()
		},
	}

	flags.Parallel(cmd)

	return cmd
}

func worker(outDir string, work <-chan job, errs chan<- error) {
	for j := range work {
		decompress(j.i, j.b, outDir, errs)
	}
}

type job struct {
	i int
	b []byte
}

func decompress(i int, b []byte, outDir string, errs chan<- error) {
	bz := bzip2.NewReader(bytes.NewReader(b))

	b, err := ioutil.ReadAll(bz)
	if err != nil {
		errs <- err
		return
	}

	err = os.MkdirAll(fmt.Sprintf("%s/extracted/%06d", outDir, i/shardSize), os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/extracted/%03d/%06d.txt", outDir, i/shardSize, i), b, os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	if i%shardSize == 0 {
		fmt.Println(i)
	}
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func extractFile(rIndex *bufio.Reader, fRepo *os.File, work chan<- job, errs chan<- error) {
	startIndex, endIndex := int64(0), int64(0)

	var (
		lineBytes []byte
		err       error
		outBytes  []byte
	)

	fileIndex := 0

	for ; err == nil; lineBytes, _, err = rIndex.ReadLine() {
		if len(lineBytes) == 0 {
			continue
		}

		if err == nil {
			line := string(lineBytes)
			parts := strings.SplitN(line, ":", 3)

			endIndex, err = strconv.ParseInt(parts[0], 10, strconv.IntSize)

			if err != nil {
				errs <- err
				return
			}

			if startIndex == endIndex {
				continue
			}

			outBytes = make([]byte, endIndex-startIndex)

			_, err = fRepo.Read(outBytes)
			if err != nil {
				errs <- err
				return
			}
		}

		work <- job{i: fileIndex, b: outBytes}
		fileIndex++

		startIndex = endIndex

		if err == io.EOF {
			break
		}
	}

	if err == io.EOF {
		fmt.Println("got last file")

		outBytes, err = ioutil.ReadAll(fRepo)
		if err != nil {
			errs <- err
			return
		}
		work <- job{i: fileIndex, b: outBytes}
	}
}
