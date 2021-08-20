package jobs

import (
	"fmt"
	"sync"
)

// Errors creates an input for errors to print and a bool of whether any
// errors have been printed.
//
// Calls Done on the WaitGroup once the channel has been closed and all errors
// have been printed.
//
// The returned channel must be closed before the WaitGroup is waited upon.
func Errors() (chan<- error, *sync.WaitGroup) {
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)

	errs := make(chan error)

	go func() {
		for err := range errs {
			fmt.Println(err)
		}

		errsWg.Done()
	}()

	return errs, &errsWg
}
