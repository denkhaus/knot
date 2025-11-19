package selection

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// CacheKey represents a unique key for cached items
type CacheKey struct {
	ProjectID     uuid.UUID
	TaskHash      string // Hash of task data for change detection
	Strategy      Strategy
	ConfigHash    string // Hash of configuration
}

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Graph         *DependencyGraph
	ComputedAt    time.Time
	ExpiresAt     time.Time
	TaskVersion   int // Simple version tracking for task changes
}

// Cache provides thread-safe caching for dependency graphs and computations
type Cache struct {
	entries map[CacheKey]*CacheEntry
	mutex   sync.RWMutex
	config  *Config
}

// NewCache creates a new cache instance
func NewCache(config *Config) *Cache {
	return &Cache{
		entries: make(map[CacheKey]*CacheEntry),
		config:  config,
	}
}

// Get retrieves a cached dependency graph if available and not expired
func (c *Cache) Get(key CacheKey) (*DependencyGraph, bool) {
	if !c.config.Advanced.CacheGraphs {
		return nil, false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Graph, true
}

// Put stores a dependency graph in the cache
func (c *Cache) Put(key CacheKey, graph *DependencyGraph) {
	if !c.config.Advanced.CacheGraphs {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := &CacheEntry{
		Graph:      graph,
		ComputedAt: time.Now(),
		ExpiresAt:  time.Now().Add(c.config.Advanced.CacheDuration),
	}

	c.entries[key] = entry
}

// Invalidate removes entries related to a specific project
func (c *Cache) Invalidate(projectID uuid.UUID) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key := range c.entries {
		if key.ProjectID == projectID {
			delete(c.entries, key)
		}
	}
}

// Cleanup removes expired entries
func (c *Cache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Size returns the number of cached entries
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[CacheKey]*CacheEntry)
}

// TaskMapCache provides caching for task maps
type TaskMapCache struct {
	entries map[string]*TaskMapEntry
	mutex   sync.RWMutex
}

// TaskMapEntry represents a cached task map
type TaskMapEntry struct {
	TaskMap    *TaskMap
	ComputedAt time.Time
	ExpiresAt  time.Time
	TaskCount  int
}

// NewTaskMapCache creates a new task map cache
func NewTaskMapCache() *TaskMapCache {
	return &TaskMapCache{
		entries: make(map[string]*TaskMapEntry),
	}
}

// GetTaskMap retrieves a cached task map
func (tmc *TaskMapCache) GetTaskMap(hash string) (*TaskMap, bool) {
	tmc.mutex.RLock()
	defer tmc.mutex.RUnlock()

	entry, exists := tmc.entries[hash]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.TaskMap, true
}

// PutTaskMap stores a task map in the cache
func (tmc *TaskMapCache) PutTaskMap(hash string, taskMap *TaskMap, duration time.Duration) {
	tmc.mutex.Lock()
	defer tmc.mutex.Unlock()

	entry := &TaskMapEntry{
		TaskMap:    taskMap,
		ComputedAt: time.Now(),
		ExpiresAt:  time.Now().Add(duration),
		TaskCount:  taskMap.Size(),
	}

	tmc.entries[hash] = entry
}

// ScoreCache provides caching for task scores
type ScoreCache struct {
	entries map[uuid.UUID]*ScoreEntry
	mutex   sync.RWMutex
}

// ScoreEntry represents a cached task score
type ScoreEntry struct {
	Score      *TaskScore
	ComputedAt time.Time
	TaskHash   string
}

// NewScoreCache creates a new score cache
func NewScoreCache() *ScoreCache {
	return &ScoreCache{
		entries: make(map[uuid.UUID]*ScoreEntry),
	}
}

// GetScore retrieves a cached task score
func (sc *ScoreCache) GetScore(taskID uuid.UUID, taskHash string) (*TaskScore, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	entry, exists := sc.entries[taskID]
	if !exists {
		return nil, false
	}

	if entry.TaskHash != taskHash {
		// Task has changed since score was computed
		return nil, false
	}

	return entry.Score, true
}

// PutScore stores a task score in the cache
func (sc *ScoreCache) PutScore(taskID uuid.UUID, score *TaskScore, taskHash string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	entry := &ScoreEntry{
		Score:      score,
		ComputedAt: score.CalculatedAt,
		TaskHash:   taskHash,
	}

	sc.entries[taskID] = entry
}