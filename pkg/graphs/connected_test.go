package graphs_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/graphs"
)

func TestFindCycle(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name  string
		graph graphs.Directed
		start uint32
		want  []uint32
	}{
		{
			name: "empty graph",
			graph: graphs.Directed{
				Nodes: nil,
			},
			start: 0,
			want:  nil,
		},
		{
			name: "singleton no edges",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "singleton cycle",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {0: true},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "two disconnected nodes",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {},
					1: {},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "two nodes one edge",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {1: true},
					1: {},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "two nodes cycle",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {1: true},
					1: {0: true},
				},
			},
			start: 0,
			want:  []uint32{0, 1},
		},
		{
			name: "two nodes cycle and self cycle",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {0: true, 1: true},
					1: {0: true},
				},
			},
			start: 0,
			want:  []uint32{0, 1},
		},
		{
			name: "long cycle",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {1: true},
					1: {2: true},
					2: {3: true},
					3: {4: true},
					4: {5: true},
					5: {6: true},
					6: {7: true},
					7: {8: true},
					8: {9: true},
					9: {0: true},
				},
			},
			start: 0,
			want:  []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "long cycle 2",
			graph: graphs.Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {1: true},
					1: {2: true},
					2: {3: true},
					3: {4: true},
					4: {5: true},
					5: {6: true},
					6: {7: true},
					7: {8: true},
					8: {9: true},
					9: {0: true},
				},
			},
			start: 2,
			want:  []uint32{2, 3, 4, 5, 6, 7, 8, 9, 0, 1},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := graphs.FindCycle(tc.start, tc.graph)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
