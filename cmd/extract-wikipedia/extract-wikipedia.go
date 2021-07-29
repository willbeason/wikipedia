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

	"github.com/spf13/cobra"
)

var cmd = cobra.Command{
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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
			err = fIndex.Close()
		}()

		rIndex := bufio.NewReader(fIndex)

		startIndex, endIndex := int64(0), int64(0)
		var lineBytes []byte
		errs := make(chan error, 100)

		go func() {
			for err = range errs {
				fmt.Println(err)
			}
		}()

		jobs := make(chan job)

		jobSync := sync.WaitGroup{}
		jobSync.Add(1)
		go func() {
			fileIndex := 0
			for ; err == nil && err != io.EOF; lineBytes, _, err = rIndex.ReadLine() {
				if len(lineBytes) == 0 {
					continue
				}

				var outBytes []byte
				if err == nil {
					line := string(lineBytes)
					parts := strings.SplitN(line, ":", 3)
					endIndex, err = strconv.ParseInt(parts[0], 10, 64)
					if err != nil {
						errs <- err
						return
					}
					if startIndex == endIndex {
						continue
					}

					outBytes = make([]byte, endIndex - startIndex)
					_, err = fRepo.Read(outBytes)
					if err != nil {
						errs <- err
						return
					}
				} else {
					fmt.Println("got last file")
					// err = io.EOF
					outBytes, err = ioutil.ReadAll(fRepo)
					if err != nil {
						errs <- err
						return
					}
				}

				jobs <- job{i: fileIndex, b: outBytes}
				fileIndex++

				startIndex = endIndex

				if err == io.EOF {
					break
				}
			}

			close(jobs)
			jobSync.Done()
		}()

		wg := sync.WaitGroup{}
		for w := 0; w < 8; w++ {
			wg.Add(1)
			go func() {
				worker(outDir, jobs, errs)
				wg.Done()
			}()
		}

		jobSync.Wait()
		wg.Wait()
		close(errs)

		return err
	},
}

func worker(outDir string, jobs <-chan job, errs chan<- error) {
	for j := range jobs {
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

	err = os.MkdirAll(fmt.Sprintf("%s/extracted/%03d", outDir, i/1000), os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/extracted/%03d/%03d.txt", outDir, i/1000, i), b, os.ModePerm)
	if err != nil {
		errs <- err
		return
	}

	if i % 1000 == 0 {
		fmt.Println(i)
	}
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
