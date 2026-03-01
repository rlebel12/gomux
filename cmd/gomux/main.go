// Package main provides the gomux entrypoint for tmux session management.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/rlebel12/gomux/internal/gomux"
	"github.com/rlebel12/gomux/internal/profile"
	"github.com/rlebel12/gomux/internal/tmux"
	"github.com/rlebel12/gomux/internal/tui"
)

const profilesConfigPath = ".config/gomux/profiles.yml"

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

	configPath := profilesConfigPath
	if home, err := os.UserHomeDir(); err == nil {
		configPath = filepath.Join(home, profilesConfigPath)
	}

	loader := profile.NewLoader(configPath)
	app := gomux.New(client, tui.NewPicker(), client, loader, client)

	if err := app.Run(ctx, sessionName); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
