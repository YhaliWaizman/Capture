package diff

import (
	"sort"

	"github.com/yhaliwaizman/capture/internal/types"
)

// DiffEngine compares declared and used variable sets
type DiffEngine interface {
	Compare(declared, used map[string]bool) types.DiffResult
}

// DiffEngineImpl implements the DiffEngine interface
type DiffEngineImpl struct{}

// NewDiffEngine creates a new DiffEngine instance
func NewDiffEngine() *DiffEngineImpl {
	return &DiffEngineImpl{}
}

// Compare performs set operations to find unused and missing variables
// Unused: declared but not used (declared - used)
// Missing: used but not declared (used - declared)
// Both results are sorted alphabetically for deterministic output
func (e *DiffEngineImpl) Compare(declared, used map[string]bool) types.DiffResult {
	var unused, missing []string

	// Find unused: in declared but not in used
	for v := range declared {
		if !used[v] {
			unused = append(unused, v)
		}
	}

	// Find missing: in used but not in declared
	for v := range used {
		if !declared[v] {
			missing = append(missing, v)
		}
	}

	// Sort for deterministic output
	sort.Strings(unused)
	sort.Strings(missing)

	return types.DiffResult{
		Unused:  unused,
		Missing: missing,
	}
}
