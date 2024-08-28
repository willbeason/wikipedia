package pagerank

type AtomicFloat64 struct {
	value float64
	deltas chan float64
}

func NewAtomicFloat() *AtomicFloat64 {
	return &AtomicFloat64{
		value:  0,
		deltas: make(chan float64, 10),
	}
}
