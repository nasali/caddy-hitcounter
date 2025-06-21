package hitcounter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/caddyserver/caddy/v2"
)

func TestHitCounter_InitialSeed(t *testing.T) {
	// Set up a temporary persistence path for this test
	tempDir := t.TempDir()
	originalPath := persistencePath
	persistencePath = filepath.Join(tempDir, "test_hitcounters.json")
	defer func() { persistencePath = originalPath }()

	tests := []struct {
		name        string
		initialSeed uint64
		key         string
		wantCount   uint64
	}{
		{
			name:        "default seed (0)",
			initialSeed: 0,
			key:         "test1",
			wantCount:   1,
		},
		{
			name:        "custom seed 100",
			initialSeed: 100,
			key:         "test2",
			wantCount:   101,
		},
		{
			name:        "custom seed 1000",
			initialSeed: 1000,
			key:         "test3",
			wantCount:   1001,
		},
		{
			name:        "large seed value",
			initialSeed: 999999,
			key:         "test4",
			wantCount:   1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HitCounter{
				InitialSeed: tt.initialSeed,
				Style:       "green",
			}

			// Create a test context with logger
			ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
			defer cancel()

			err := hc.Provision(ctx)
			if err != nil {
				t.Fatalf("Provision failed: %v", err)
			}

			// Get the template function
			funcMap := hc.CustomTemplateFunctions()
			hitCounterFunc, ok := funcMap["hitCounter"].(func(string) (string, error))
			if !ok {
				t.Fatal("hitCounter function should exist")
			}

			// Call the function
			result, err := hitCounterFunc(tt.key)
			if err != nil {
				t.Fatalf("hitCounter function failed: %v", err)
			}
			if result == "" {
				t.Fatal("result should not be empty")
			}

			// Verify the counter was incremented from the seed
			hc.countersMu.Lock()
			count := hc.counters[tt.key]
			hc.countersMu.Unlock()
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestHitCounter_MultipleIncrements(t *testing.T) {
	// Set up a temporary persistence path for this test
	tempDir := t.TempDir()
	originalPath := persistencePath
	persistencePath = filepath.Join(tempDir, "test_hitcounters.json")
	defer func() { persistencePath = originalPath }()

	hc := &HitCounter{
		InitialSeed: 50,
		Style:       "green",
	}

	// Create a test context with logger
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	err := hc.Provision(ctx)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	funcMap := hc.CustomTemplateFunctions()
	hitCounterFunc, ok := funcMap["hitCounter"].(func(string) (string, error))
	if !ok {
		t.Fatal("hitCounter function should exist")
	}

	// First increment
	_, err = hitCounterFunc("counter1")
	if err != nil {
		t.Fatalf("hitCounter function failed: %v", err)
	}

	hc.countersMu.Lock()
	count := hc.counters["counter1"]
	hc.countersMu.Unlock()
	if count != uint64(51) {
		t.Errorf("count = %d, want %d", count, 51)
	}

	// Second increment
	_, err = hitCounterFunc("counter1")
	if err != nil {
		t.Fatalf("hitCounter function failed: %v", err)
	}

	hc.countersMu.Lock()
	count = hc.counters["counter1"]
	hc.countersMu.Unlock()
	if count != uint64(52) {
		t.Errorf("count = %d, want %d", count, 52)
	}

	// Different counter with same seed
	_, err = hitCounterFunc("counter2")
	if err != nil {
		t.Fatalf("hitCounter function failed: %v", err)
	}

	hc.countersMu.Lock()
	count = hc.counters["counter2"]
	hc.countersMu.Unlock()
	if count != uint64(51) {
		t.Errorf("count = %d, want %d", count, 51)
	}
}

func TestHitCounter_PaddingWithSeed(t *testing.T) {
	// Set up a temporary persistence path for this test
	tempDir := t.TempDir()
	originalPath := persistencePath
	persistencePath = filepath.Join(tempDir, "test_hitcounters.json")
	defer func() { persistencePath = originalPath }()

	hc := &HitCounter{
		InitialSeed: 98,
		PadDigits:   4,
		Style:       "green",
	}

	// Create a test context with logger
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	err := hc.Provision(ctx)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	funcMap := hc.CustomTemplateFunctions()
	hitCounterFunc, ok := funcMap["hitCounter"].(func(string) (string, error))
	if !ok {
		t.Fatal("hitCounter function should exist")
	}

	// Should increment to 99
	result, err := hitCounterFunc("padtest")
	if err != nil {
		t.Fatalf("hitCounter function failed: %v", err)
	}
	// Check that result contains 4 img tags (for 0099)
	if !strings.Contains(result, "<img") {
		t.Error("result should contain <img tag")
	}
	if !strings.Contains(result, `alt="0"`) {
		t.Error("result should contain alt=\"0\"")
	}
	if !strings.Contains(result, `alt="9"`) {
		t.Error("result should contain alt=\"9\"")
	}
}

func TestHitCounter_TemplateFunctionIntegration(t *testing.T) {
	// Set up a temporary persistence path for this test
	tempDir := t.TempDir()
	originalPath := persistencePath
	persistencePath = filepath.Join(tempDir, "test_hitcounters.json")
	defer func() { persistencePath = originalPath }()

	hc := &HitCounter{
		InitialSeed: 500,
		Style:       "green",
	}

	// Create a test context with logger
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	err := hc.Provision(ctx)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	// Test that the function can be used in a template
	tmpl := template.New("test")
	tmpl.Funcs(hc.CustomTemplateFunctions())

	parsed, err := tmpl.Parse(`{{ hitCounter "page1" }}`)
	if err != nil {
		t.Fatalf("template parse failed: %v", err)
	}

	var buf strings.Builder
	err = parsed.Execute(&buf, nil)
	if err != nil {
		t.Fatalf("template execute failed: %v", err)
	}

	result := buf.String()
	if result == "" {
		t.Error("result should not be empty")
	}
	if !strings.Contains(result, "<img") {
		t.Error("result should contain <img tag")
	}
	if !strings.Contains(result, "Hit counter") {
		t.Error("result should contain 'Hit counter'")
	}
}

func testContext(t *testing.T) caddy.Context {
	t.Helper()
	ctx, _ := caddy.NewContext(caddy.Context{Context: context.Background()})
	return ctx
}