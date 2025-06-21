package hitcounter

import (
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStyle string
		expectedPad   int
		expectedSeed  uint64
		expectError   bool
	}{
		{
			name: "empty config",
			input: `hitCounter {
			}`,
			expectedStyle: "",
			expectedPad:   0,
			expectedSeed:  0,
			expectError:   false,
		},
		{
			name: "style only",
			input: `hitCounter {
				style green
			}`,
			expectedStyle: "green",
			expectedPad:   0,
			expectedSeed:  0,
			expectError:   false,
		},
		{
			name: "pad_digits only",
			input: `hitCounter {
				pad_digits 5
			}`,
			expectedStyle: "",
			expectedPad:   5,
			expectedSeed:  0,
			expectError:   false,
		},
		{
			name: "initial_seed only",
			input: `hitCounter {
				initial_seed 1000
			}`,
			expectedStyle: "",
			expectedPad:   0,
			expectedSeed:  1000,
			expectError:   false,
		},
		{
			name: "all options",
			input: `hitCounter {
				style odometer
				pad_digits 7
				initial_seed 50000
			}`,
			expectedStyle: "odometer",
			expectedPad:   7,
			expectedSeed:  50000,
			expectError:   false,
		},
		{
			name: "large initial_seed",
			input: `hitCounter {
				initial_seed 9999999999
			}`,
			expectedStyle: "",
			expectedPad:   0,
			expectedSeed:  9999999999,
			expectError:   false,
		},
		{
			name: "invalid initial_seed",
			input: `hitCounter {
				initial_seed abc
			}`,
			expectError: true,
		},
		{
			name: "negative initial_seed",
			input: `hitCounter {
				initial_seed -100
			}`,
			expectError: true,
		},
		{
			name: "invalid pad_digits",
			input: `hitCounter {
				pad_digits xyz
			}`,
			expectError: true,
		},
		{
			name: "missing style argument",
			input: `hitCounter {
				style
			}`,
			expectError: true,
		},
		{
			name: "missing initial_seed argument",
			input: `hitCounter {
				initial_seed
			}`,
			expectError: true,
		},
		{
			name: "too many arguments for initial_seed",
			input: `hitCounter {
				initial_seed 100 200
			}`,
			expectError: true,
		},
		{
			name: "unknown property",
			input: `hitCounter {
				unknown_property value
			}`,
			expectError: true,
		},
		{
			name: "multiple styles",
			input: `hitCounter {
				style green
				style yellow
			}`,
			expectedStyle: "yellow",
			expectedPad:   0,
			expectedSeed:  0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HitCounter{}
			d := caddyfile.NewTestDispenser(tt.input)

			err := hc.UnmarshalCaddyfile(d)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if hc.Style != tt.expectedStyle {
					t.Errorf("Style = %q, want %q", hc.Style, tt.expectedStyle)
				}
				if hc.PadDigits != tt.expectedPad {
					t.Errorf("PadDigits = %d, want %d", hc.PadDigits, tt.expectedPad)
				}
				if hc.InitialSeed != tt.expectedSeed {
					t.Errorf("InitialSeed = %d, want %d", hc.InitialSeed, tt.expectedSeed)
				}
			}
		})
	}
}

func TestUnmarshalCaddyfile_EdgeCases(t *testing.T) {
	t.Run("max uint64 seed", func(t *testing.T) {
		hc := &HitCounter{}
		input := `hitCounter {
			initial_seed 18446744073709551615
		}`
		d := caddyfile.NewTestDispenser(input)
		err := hc.UnmarshalCaddyfile(d)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hc.InitialSeed != uint64(18446744073709551615) {
			t.Errorf("InitialSeed = %d, want %d", hc.InitialSeed, uint64(18446744073709551615))
		}
	})

	t.Run("overflow uint64 seed", func(t *testing.T) {
		hc := &HitCounter{}
		input := `hitCounter {
			initial_seed 18446744073709551616
		}`
		d := caddyfile.NewTestDispenser(input)
		err := hc.UnmarshalCaddyfile(d)
		if err == nil {
			t.Error("expected error but got none")
		}
		if !strings.Contains(err.Error(), "invalid initial seed value") {
			t.Errorf("error message should contain 'invalid initial seed value', got: %v", err)
		}
	})

	t.Run("zero seed explicitly set", func(t *testing.T) {
		hc := &HitCounter{}
		input := `hitCounter {
			initial_seed 0
		}`
		d := caddyfile.NewTestDispenser(input)
		err := hc.UnmarshalCaddyfile(d)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hc.InitialSeed != uint64(0) {
			t.Errorf("InitialSeed = %d, want %d", hc.InitialSeed, 0)
		}
	})
}