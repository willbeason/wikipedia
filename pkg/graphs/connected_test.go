package graphs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindCycle(t *testing.T) {
	tcs := []struct {
		name  string
		graph Directed
		start uint32
		want  []uint32
	}{
		{
			name:  "empty graph",
			graph: Directed{},
			start: 0,
			want:  nil,
		},
		{
			name: "singleton no edges",
			graph: Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "singleton cycle",
			graph: Directed{
				Nodes: map[uint32]map[uint32]bool{
					0: {0: true},
				},
			},
			start: 0,
			want:  nil,
		},
		{
			name: "two disconnected nodes",
			graph: Directed{
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
			graph: Directed{
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
			graph: Directed{
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
			graph: Directed{
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
			graph: Directed{
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
			graph: Directed{
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
			got := FindCycle(tc.start, tc.graph)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
