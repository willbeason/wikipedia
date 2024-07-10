package charts_test

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/charts"
)

func TestLogarithmicBins(t *testing.T) {
	t.Parallel()

	want := []int{
		10, 22, 46,
		100, 215, 464,
		1000, 2154, 4642,
		10000, 21544, 46416,
		100000, 215443, 464159,
		1000000, 2154435, 4641589,
	}

	got := charts.LogarithmicBins(10, 10000000, math.Pow(10.0, 1.0/3.0))

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("LogarithmicBins() mismatch (-want +got):\n%s", diff)
	}
}
