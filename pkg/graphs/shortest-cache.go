package graphs

import "sync"

type FromTo struct {
	From uint32
	To   uint32
}

type ShortestCache struct {
	shards   uint32
	locks    []sync.RWMutex
	distance []map[FromTo]int
}

func NewShortestCache(shards uint32) *ShortestCache {
	result := &ShortestCache{
		shards:   shards,
		locks:    make([]sync.RWMutex, shards),
		distance: make([]map[FromTo]int, shards),
	}

	for i := uint32(0); i < shards; i++ {
		result.distance[i] = make(map[FromTo]int)
	}

	return result
}

func (c *ShortestCache) Get(from, to uint32) (int, bool) {
	ft := FromTo{From: from, To: to}

	shard := from % c.shards

	c.locks[shard].RLock()
	d, ok := c.distance[shard][ft]
	c.locks[shard].RUnlock()

	return d, ok
}

func (c *ShortestCache) Add(from, to uint32, dist int) {
	ft := FromTo{From: from, To: to}

	shard := from % c.shards

	c.locks[shard].RLock()
	_, exists := c.distance[shard][ft]
	c.locks[shard].RUnlock()

	if exists {
		return
	}

	c.locks[shard].Lock()
	d := c.distance[shard]
	d[ft] = dist
	c.distance[shard] = d
	c.locks[shard].Unlock()
}

func (c *ShortestCache) AddNext(from, to, next uint32) {
	dist, exists := c.Get(from, to)

	if !exists {
		return
	}

	c.Add(from, next, dist+1)
}
