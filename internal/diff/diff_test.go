package diff

import (
	"reflect"
	"testing"
)

func TestDiffEngine_Compare_EmptySets(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{}
	used := map[string]bool{}

	result := engine.Compare(declared, used)

	if len(result.Unused) != 0 {
		t.Errorf("Expected no unused variables, got %v", result.Unused)
	}
	if len(result.Missing) != 0 {
		t.Errorf("Expected no missing variables, got %v", result.Missing)
	}
}

func TestDiffEngine_Compare_IdenticalSets(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"PORT":         true,
	}
	used := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"PORT":         true,
	}

	result := engine.Compare(declared, used)

	if len(result.Unused) != 0 {
		t.Errorf("Expected no unused variables, got %v", result.Unused)
	}
	if len(result.Missing) != 0 {
		t.Errorf("Expected no missing variables, got %v", result.Missing)
	}
}

func TestDiffEngine_Compare_OnlyUnused(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"UNUSED_VAR":   true,
	}
	used := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
	}

	result := engine.Compare(declared, used)

	expected := []string{"UNUSED_VAR"}
	if !reflect.DeepEqual(result.Unused, expected) {
		t.Errorf("Expected unused %v, got %v", expected, result.Unused)
	}
	if len(result.Missing) != 0 {
		t.Errorf("Expected no missing variables, got %v", result.Missing)
	}
}

func TestDiffEngine_Compare_OnlyMissing(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
	}
	used := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"MISSING_VAR":  true,
	}

	result := engine.Compare(declared, used)

	if len(result.Unused) != 0 {
		t.Errorf("Expected no unused variables, got %v", result.Unused)
	}
	expected := []string{"MISSING_VAR"}
	if !reflect.DeepEqual(result.Missing, expected) {
		t.Errorf("Expected missing %v, got %v", expected, result.Missing)
	}
}

func TestDiffEngine_Compare_BothUnusedAndMissing(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":    true,
		"UNUSED_VAR": true,
	}
	used := map[string]bool{
		"API_KEY":     true,
		"MISSING_VAR": true,
	}

	result := engine.Compare(declared, used)

	expectedUnused := []string{"UNUSED_VAR"}
	if !reflect.DeepEqual(result.Unused, expectedUnused) {
		t.Errorf("Expected unused %v, got %v", expectedUnused, result.Unused)
	}
	expectedMissing := []string{"MISSING_VAR"}
	if !reflect.DeepEqual(result.Missing, expectedMissing) {
		t.Errorf("Expected missing %v, got %v", expectedMissing, result.Missing)
	}
}

func TestDiffEngine_Compare_AlphabeticalSorting(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"ZEBRA":   true,
		"ALPHA":   true,
		"CHARLIE": true,
		"BRAVO":   true,
	}
	used := map[string]bool{
		"YANKEE": true,
		"XRAY":   true,
		"ZULU":   true,
	}

	result := engine.Compare(declared, used)

	// Check unused is sorted alphabetically
	expectedUnused := []string{"ALPHA", "BRAVO", "CHARLIE", "ZEBRA"}
	if !reflect.DeepEqual(result.Unused, expectedUnused) {
		t.Errorf("Expected unused sorted as %v, got %v", expectedUnused, result.Unused)
	}

	// Check missing is sorted alphabetically
	expectedMissing := []string{"XRAY", "YANKEE", "ZULU"}
	if !reflect.DeepEqual(result.Missing, expectedMissing) {
		t.Errorf("Expected missing sorted as %v, got %v", expectedMissing, result.Missing)
	}
}

func TestDiffEngine_Compare_Idempotence(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":    true,
		"UNUSED_VAR": true,
	}
	used := map[string]bool{
		"API_KEY":     true,
		"MISSING_VAR": true,
	}

	// Run comparison twice
	result1 := engine.Compare(declared, used)
	result2 := engine.Compare(declared, used)

	// Results should be identical
	if !reflect.DeepEqual(result1, result2) {
		t.Errorf("Expected idempotent results, got different results:\n%v\n%v", result1, result2)
	}
}

func TestDiffEngine_Compare_MultipleUnusedAndMissing(t *testing.T) {
	engine := NewDiffEngine()
	declared := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"UNUSED_ONE":   true,
		"UNUSED_TWO":   true,
		"UNUSED_THREE": true,
	}
	used := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"MISSING_ONE":  true,
		"MISSING_TWO":  true,
	}

	result := engine.Compare(declared, used)

	expectedUnused := []string{"UNUSED_ONE", "UNUSED_THREE", "UNUSED_TWO"}
	if !reflect.DeepEqual(result.Unused, expectedUnused) {
		t.Errorf("Expected unused %v, got %v", expectedUnused, result.Unused)
	}

	expectedMissing := []string{"MISSING_ONE", "MISSING_TWO"}
	if !reflect.DeepEqual(result.Missing, expectedMissing) {
		t.Errorf("Expected missing %v, got %v", expectedMissing, result.Missing)
	}
}
