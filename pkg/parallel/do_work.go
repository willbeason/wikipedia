package parallel

import "fmt"

func DoWork(work <-chan string, doWork func(string) error, errs chan<- error) {
	for item := range work {
		err := doWork(item)
		if err != nil {
			errs <- fmt.Errorf("item: %w", err)
		}
	}
}
