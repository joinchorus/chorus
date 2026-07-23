package conversationname

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Generator defines the contract for producing thread-scoped conversation names.
type Generator interface {
	Generate(usedNames []string) string
	RandomName() string
}

// DefaultGenerator implements Generator using a thread-safe word pool.
type DefaultGenerator struct {
	mu    sync.Mutex
	rng   *rand.Rand
	words []string
}

// NewDefaultGenerator constructs a new DefaultGenerator.
// If words is empty or nil, DefaultWordList is used.
func NewDefaultGenerator(words []string) *DefaultGenerator {
	if len(words) == 0 {
		words = DefaultWordList
	}
	// Copy slice to avoid external mutation
	wordCopy := make([]string, len(words))
	copy(wordCopy, words)

	src := rand.NewSource(time.Now().UnixNano())
	return &DefaultGenerator{
		rng:   rand.New(src),
		words: wordCopy,
	}
}

// RandomName returns a random name from the pool without considering used names.
func (g *DefaultGenerator) RandomName() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	idx := g.rng.Intn(len(g.words))
	return g.words[idx]
}

// Generate produces a conversation name that is guaranteed to be unique within a thread.
// It checks usedNames (case-insensitive) and selects an unused name from the pool.
// If all base names are exhausted, it appends a numeric suffix (e.g., "Ash 2").
func (g *DefaultGenerator) Generate(usedNames []string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	usedMap := make(map[string]bool, len(usedNames))
	for _, n := range usedNames {
		usedMap[strings.ToLower(strings.TrimSpace(n))] = true
	}

	// Collect available candidates
	available := make([]string, 0, len(g.words))
	for _, w := range g.words {
		if !usedMap[strings.ToLower(w)] {
			available = append(available, w)
		}
	}

	if len(available) > 0 {
		idx := g.rng.Intn(len(available))
		return available[idx]
	}

	// Fallback when base word pool is completely exhausted for a thread
	// Pick a random base word and increment numeric suffix until unique
	baseIdx := g.rng.Intn(len(g.words))
	baseName := g.words[baseIdx]
	counter := 2
	for {
		candidate := fmt.Sprintf("%s %d", baseName, counter)
		if !usedMap[strings.ToLower(candidate)] {
			return candidate
		}
		counter++
	}
}
