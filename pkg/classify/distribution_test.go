package classify_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/willbeason/wikipedia/pkg/classify"
)

func TestLogDistribution_ToDistribution(t *testing.T) {
	tcs := []struct {
		name            string
		logDistribution classify.LogDistribution
		want            classify.Distribution
	}{
		{
			name:            "nil distribution",
			logDistribution: nil,
			want:            classify.Distribution{},
		},
		{
			name:            "empty distribution",
			logDistribution: classify.LogDistribution{},
			want:            classify.Distribution{},
		},
		{
			name:            "one entry",
			logDistribution: classify.LogDistribution{0.0},
			want:            classify.Distribution{1.0},
		},
		{
			name:            "one entry very negative",
			logDistribution: classify.LogDistribution{-1000.0},
			want:            classify.Distribution{1.0},
		},
		{
			name:            "three entries",
			logDistribution: classify.LogDistribution{0.0, -1.0, -2.0},
			want:            classify.Distribution{0.665, 0.244, 0.090},
		},
		{
			name:            "three entries offset",
			logDistribution: classify.LogDistribution{-100.0, -101.0, -102.0},
			want:            classify.Distribution{0.665, 0.244, 0.090},
		},
		{
			name:            "extreme differences",
			logDistribution: classify.LogDistribution{0.0, -10.0, -20.0},
			want:            classify.Distribution{1.000, 0.000, 0.000},
		},
		{
			name:            "extreme differences",
			logDistribution: classify.LogDistribution{0.0, -1000.0, -2000.0},
			want:            classify.Distribution{1.000, 0.000, 0.000},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.logDistribution.ToDistribution()

			if diff := cmp.Diff(tc.want, got, cmpopts.EquateApprox(0.0, 0.001)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestDistribution_Normalize(t *testing.T) {
	tcs := []struct {
		name   string
		before classify.Distribution
		want   classify.Distribution
	}{
		{
			name:   "nil distribution",
			before: nil,
			want:   nil,
		},
		{
			name:   "empty distribution",
			before: classify.Distribution{},
			want:   classify.Distribution{},
		},
		{
			name:   "zero distribution one entry",
			before: classify.Distribution{0.0},
			want:   classify.Distribution{0.0},
		},
		{
			name:   "zero distribution three entries",
			before: classify.Distribution{0.0, 0.0, 0.0},
			want:   classify.Distribution{0.0, 0.0, 0.0},
		},
		{
			name:   "identity distribution",
			before: classify.Distribution{1.0},
			want:   classify.Distribution{1.0},
		},
		{
			name:   "one small element",
			before: classify.Distribution{0.01},
			want:   classify.Distribution{1.0},
		},
		{
			name:   "three equal elements",
			before: classify.Distribution{1.0, 1.0, 1.0},
			want:   classify.Distribution{0.333, 0.333, 0.333},
		},
		{
			name:   "three unequal elements",
			before: classify.Distribution{3.0, 2.0, 1.0},
			want:   classify.Distribution{0.5, 0.333, 0.167},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.before.Normalize()

			if diff := cmp.Diff(tc.want, tc.before, cmpopts.EquateApprox(0.0, 0.001)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
