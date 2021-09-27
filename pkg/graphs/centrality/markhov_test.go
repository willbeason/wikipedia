package centrality

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/willbeason/wikipedia/pkg/graphs"
)

func TestMarkhov(t *testing.T) {
	tcs := []struct {
		name  string
		graph *graphs.Directed
		want  map[uint32]float64
	}{
		{
			name: "two squares",
			graph: &graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					1: {2: true, 3: true},
					2: {3: true},
					3: {4: true, 5: true},
					4: {5: true},
					5: {6: true, 7: true},
					6: {7: true},
					7: {8: true, 1: true},
					8: {1: true},
				},
			},
			want: map[uint32]float64{
				1: 4.0 / 3.0,
				2: 2.0 / 3.0,
				3: 4.0 / 3.0,
				4: 2.0 / 3.0,
				5: 4.0 / 3.0,
				6: 2.0 / 3.0,
				7: 4.0 / 3.0,
				8: 2.0 / 3.0,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Markhov(tc.graph, 0.001, 100)

			if diff := cmp.Diff(tc.want, got, cmpopts.EquateApprox(0.001, 0.0)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
