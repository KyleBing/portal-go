package setup

import (
	"os"
	"testing"
)

func TestAllowSetupEnv(t *testing.T) {
	orig := os.Getenv("ALLOW_SETUP")
	t.Cleanup(func() { _ = os.Setenv("ALLOW_SETUP", orig) })

	cases := []struct {
		val  string
		want bool
	}{
		{"", true},
		{"1", true},
		{"true", true},
		{"0", false},
		{"false", false},
		{"OFF", false},
		{"no", false},
	}
	for _, tc := range cases {
		_ = os.Setenv("ALLOW_SETUP", tc.val)
		if got := AllowSetupEnv(); got != tc.want {
			t.Fatalf("ALLOW_SETUP=%q: got %v want %v", tc.val, got, tc.want)
		}
	}
}
