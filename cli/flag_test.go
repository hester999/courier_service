package cli

import (
	"flag"
	"os"
	"testing"

	"avito/pkg"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestSetFlags_PortNotProvided(t *testing.T) {
	resetFlags()
	logger := pkg.NewLogger(0)

	t.Setenv("PORT", "")

	os.Args = []string{"cmd"}

	cli := NewFlagsCLI(logger)

	if err := cli.SetFlags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("PORT"); got != "" {
		t.Fatalf("PORT should not be set, got %q", got)
	}
}

func TestSetFlags_PortProvided(t *testing.T) {
	resetFlags()
	t.Setenv("PORT", "")
	logger := pkg.NewLogger(0)
	os.Args = []string{"cmd", "-port=8080"}

	cli := NewFlagsCLI(logger)

	if err := cli.SetFlags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("PORT"); got != "8080" {
		t.Fatalf("expected PORT=8080, got %q", got)
	}
}

func TestSetFlags_OverrideEnv(t *testing.T) {
	resetFlags()
	t.Setenv("PORT", "3000")
	logger := pkg.NewLogger(0)
	os.Args = []string{"cmd", "-port=9090"}

	cli := NewFlagsCLI(logger)

	if err := cli.SetFlags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("PORT"); got != "9090" {
		t.Fatalf("expected PORT=9090, got %q", got)
	}
}
