// Package main provides the gomux entrypoint for tmux session management.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/rlebel12/gomux/internal/gomux"
	"github.com/rlebel12/gomux/internal/tmux"
	"github.com/rlebel12/gomux/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gomux: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := tmux.NewClient()
	if err != nil {
		return fmt.Errorf("create tmux client: %w", err)
	}

	sessionName := ""
	if len(os.Args) > 1 {
		sessionName = os.Args[1]
	}

	// TODO(phase-04): replace legacyPicker with updated tui.Picker once Phase 04 is complete.
	// TODO(phase-06): replace noopProfileLoader and noopSessionCreator with real implementations.
	app := gomux.New(client, &legacyPicker{tui.NewPicker()}, client, noopProfileLoader{}, noopSessionCreator{})

	if err := app.Run(ctx, sessionName); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}

// legacyPicker adapts the existing tui.Picker (old signature) to the updated SessionPicker interface.
// TODO(phase-04): remove once tui.Picker is updated in Phase 04.
type legacyPicker struct {
	inner *tui.Picker
}

func (lp *legacyPicker) Pick(ctx context.Context, sessions []gomux.Session, _ []gomux.Profile) (gomux.Selection, error) {
	name, err := lp.inner.Pick(ctx, sessions)
	if err != nil {
		return gomux.Selection{}, err
	}

	return gomux.Selection{Name: name, IsProfile: false}, nil
}

// noopProfileLoader is a temporary stub until Phase 06 wires the real ProfileLoader.
type noopProfileLoader struct{}

func (noopProfileLoader) LoadProfiles(_ context.Context) ([]gomux.Profile, error) {
	return nil, nil
}

// noopSessionCreator is a temporary stub until Phase 06 wires the real SessionCreator.
type noopSessionCreator struct{}

func (noopSessionCreator) CreateSessionFromProfile(_ context.Context, _ gomux.Profile) error {
	return nil
}
