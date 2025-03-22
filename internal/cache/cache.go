package cache

import (
	"container/list"
	"sync"
)

type EpisodeCache struct {
	mu       sync.Mutex
	maxSize  int
	cache    map[string]*list.Element
	eviction *list.List
}

type entry struct {
	key   string
	value map[int]int
}

func CreateEpisodeCache(maxSize int) *EpisodeCache {
	return &EpisodeCache{
		maxSize:  maxSize,
		cache:    make(map[string]*list.Element),
		eviction: list.New(),
	}
}

func (c *EpisodeCache) Get(imdbID string, season int) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.cache[imdbID]; found {
		c.eviction.MoveToFront(elem)
		seasons := elem.Value.(*entry).value
		episodeCount, exists := seasons[season]
		return episodeCount, exists
	}
	return 0, false
}

func (c *EpisodeCache) Set(imdbID string, season, episodeCount int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.cache[imdbID]; found {
		elem.Value.(*entry).value[season] = episodeCount
		c.eviction.MoveToFront(elem)
		return
	}

	if len(c.cache) >= c.maxSize {
		oldest := c.eviction.Back()
		if oldest != nil {
			delete(c.cache, oldest.Value.(*entry).key)
			c.eviction.Remove(oldest)
		}
	}

	newEntry := &entry{
		key:   imdbID,
		value: map[int]int{season: episodeCount},
	}
	elem := c.eviction.PushFront(newEntry)
	c.cache[imdbID] = elem
}
