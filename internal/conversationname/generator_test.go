package conversationname

import (
	"strings"
	"testing"
)

func TestDefaultGenerator_RandomName(t *testing.T) {
	gen := NewDefaultGenerator(nil)
	name := gen.RandomName()
	if name == "" {
		t.Fatalf("expected non-empty random name")
	}
}

func TestDefaultGenerator_ThreadUniqueness(t *testing.T) {
	pool := []string{"Ash", "River", "Echo"}
	gen := NewDefaultGenerator(pool)

	used := []string{"Ash"}
	generated := gen.Generate(used)

	if strings.EqualFold(generated, "Ash") {
		t.Fatalf("expected name other than Ash, got %s", generated)
	}

	if generated != "River" && generated != "Echo" {
		t.Fatalf("expected River or Echo, got %s", generated)
	}
}

func TestDefaultGenerator_ThreadIndependence(t *testing.T) {
	pool := []string{"Ash", "River", "Echo"}
	gen := NewDefaultGenerator(pool)

	// Thread A has used Ash and River
	usedThreadA := []string{"Ash", "River"}
	nameA := gen.Generate(usedThreadA)
	if nameA != "Echo" {
		t.Fatalf("expected Echo for Thread A, got %s", nameA)
	}

	// Thread B has used Echo only -> Ash or River is valid for Thread B
	usedThreadB := []string{"Echo"}
	nameB := gen.Generate(usedThreadB)
	if nameB == "Echo" {
		t.Fatalf("expected Ash or River for Thread B, got %s", nameB)
	}
}

func TestDefaultGenerator_ExhaustedPoolFallback(t *testing.T) {
	pool := []string{"Ash"}
	gen := NewDefaultGenerator(pool)

	used := []string{"Ash"}
	firstFallback := gen.Generate(used)
	if firstFallback != "Ash 2" {
		t.Fatalf("expected Ash 2, got %s", firstFallback)
	}

	used = append(used, "Ash 2")
	secondFallback := gen.Generate(used)
	if secondFallback != "Ash 3" {
		t.Fatalf("expected Ash 3, got %s", secondFallback)
	}
}

func TestDefaultGenerator_NoDuplicateNamesInThread(t *testing.T) {
	pool := []string{"Ash", "River", "Echo", "Stone", "Willow"}
	gen := NewDefaultGenerator(pool)

	used := make([]string, 0)
	seen := make(map[string]bool)

	for i := 0; i < len(pool); i++ {
		name := gen.Generate(used)
		if seen[strings.ToLower(name)] {
			t.Fatalf("duplicate name generated: %s", name)
		}
		seen[strings.ToLower(name)] = true
		used = append(used, name)
	}

	if len(used) != len(pool) {
		t.Fatalf("expected %d unique names, got %d", len(pool), len(used))
	}
}
