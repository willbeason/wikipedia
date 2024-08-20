package documents

import (
	"errors"
	"fmt"
)

var ErrLoop = errors.New("redirect loop detected")

func GetDestination(redirects *Redirects, titleIndex map[string]uint32, title string) (string, error) {
	if _, notRedirect := titleIndex[title]; notRedirect {
		// This isn't actually a redirect since the article really exists.
		return title, nil
	}

	destination, isRedirect := redirects.Redirects[title]
	if title == destination {
		// Self redirect.
		return title, nil
	}

	seen := map[string]bool{
		title: true,
	}
	loop := []string{title}

	for ; isRedirect; destination, isRedirect = redirects.Redirects[title] {
		loop = append(loop, destination)

		if seen[destination] {
			return "", fmt.Errorf("%w: %v", ErrLoop, loop)
		}

		seen[destination] = true

		title = destination
		if _, notRedirect := titleIndex[title]; notRedirect {
			// This isn't actually a redirect since the article really exists.
			return title, nil
		}
	}

	return title, nil
}
