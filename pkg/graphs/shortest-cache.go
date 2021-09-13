package graphs

import "sync"

type FromTo struct {
	From uint32
	To   uint32
}

type ShortestCache struct {
	mtx      sync.RWMutex
	Distance map[FromTo]int
}

func (c *ShortestCache) Get(from, to uint32) (int, bool) {
	ft := FromTo{From: from, To: to}

	c.mtx.RLock()
	d, ok := c.Distance[ft]
	c.mtx.RUnlock()

	return d, ok
}

func (c *ShortestCache) Add(from, to uint32, dist int) {
	ft := FromTo{From: from, To: to}

	c.mtx.RLock()
	_, exists := c.Distance[ft]
	c.mtx.RUnlock()

	if exists {
		return
	}

	c.mtx.Lock()
	c.Distance[ft] = dist
	c.mtx.Unlock()
}

func (c *ShortestCache) AddNext(from, to, next uint32) {
	dist, exists := c.Get(from, to)

	if !exists {
		return
	}

	c.Add(from, next, dist+1)
}
