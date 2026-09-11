package operations_test

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCreate_ConcurrentRequestsNeverCollide(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}
	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}

	const n = 6
	results := make([]operations.CreateResult, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = operations.Create(root, cfg, operations.CreateRequest{
				Type:   ids.Spec,
				Parent: feat.ID.String(),
			})
		}(i)
	}
	wg.Wait()

	seen := map[int]bool{}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("Create() call %d unexpected error: %v", i, err)
		}
		num := results[i].ID.Number
		if seen[num] {
			t.Fatalf("duplicate Spec number %d allocated by two concurrent Create() calls", num)
		}
		seen[num] = true

		// Every result's artifact must be fully, correctly written —
		// never corrupted or partial.
		if _, err := operations.Inspect(root, cfg, results[i].ID.String()); err != nil {
			t.Fatalf("Inspect(%v) after concurrent Create() unexpected error: %v", results[i].ID, err)
		}
	}

	var numbers []int
	for num := range seen {
		numbers = append(numbers, num)
	}
	sort.Ints(numbers)
	for i, num := range numbers {
		if num != i+1 {
			t.Fatalf("allocated numbers = %v, want a contiguous 1..%d sequence with no gaps", numbers, n)
		}
	}
}

func TestCreate_RecoversFromStaleLockAutomatically(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	lockPath := filepath.Join(root, ".misterspec", ".lock")
	if err := os.WriteFile(lockPath, []byte("leftover from a crashed process\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// No manual cleanup of lockPath here — Create must recover it itself.
	result, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create() unexpected error after a stale lock was present: %v", err)
	}
	if result.ID.String() != "PRG-001" {
		t.Errorf("ID = %v, want PRG-001", result.ID)
	}
}
