package gomux_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rlebel12/gomux/internal/gomux"
)

var (
	errTmuxServerNotRunning = errors.New("tmux server not running")
	errTmuxBinaryNotFound   = errors.New("tmux binary not found")
)

// Test fakes

type fakeLister struct {
	err      error
	sessions []gomux.Session
	called   bool
}

func (f *fakeLister) ListSessions(_ context.Context) ([]gomux.Session, error) {
	f.called = true
	return f.sessions, f.err
}

type fakePicker struct {
	err      error
	received []gomux.Session
	name     string
	called   bool
}

func (f *fakePicker) Pick(_ context.Context, sessions []gomux.Session) (string, error) {
	f.called = true
	f.received = sessions

	return f.name, f.err
}

type fakeExecutor struct {
	err         error
	sessionName string
	called      bool
}

func (f *fakeExecutor) ExecIntoSession(_ context.Context, name string) error {
	f.called = true
	f.sessionName = name

	return f.err
}

// Parameterized tests for App.Run routing logic

func TestAppRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		listerErr        error
		pickerErr        error
		executorErr      error
		listerSessions   []gomux.Session
		wantErr          bool
		wantListerCall   bool
		wantPickerCall   bool
		wantExecutorCall bool
		name             string
		sessionName      string
		pickerName       string
		wantExecName     string
	}{
		{
			name:             "direct_mode",
			sessionName:      "work",
			wantExecName:     "work",
			wantListerCall:   false,
			wantPickerCall:   false,
			wantExecutorCall: true,
		},
		{
			name:             "empty_state",
			sessionName:      "",
			listerSessions:   []gomux.Session{},
			wantExecName:     gomux.DefaultSessionName,
			wantListerCall:   true,
			wantPickerCall:   false,
			wantExecutorCall: true,
		},
		{
			name:        "tui_mode_single",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 1, Created: time.Now()},
			},
			pickerName:       "dev",
			wantExecName:     "dev",
			wantListerCall:   true,
			wantPickerCall:   true,
			wantExecutorCall: true,
		},
		{
			name:        "tui_mode_multi",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 2, Created: time.Now()},
				{Name: "api", Windows: 1, Created: time.Now()},
			},
			pickerName:       "api",
			wantExecName:     "api",
			wantListerCall:   true,
			wantPickerCall:   true,
			wantExecutorCall: true,
		},
		{
			name:        "user_quit",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 1, Created: time.Now()},
			},
			pickerErr:        gomux.ErrNoSelection,
			wantListerCall:   true,
			wantPickerCall:   true,
			wantExecutorCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lister := &fakeLister{sessions: tt.listerSessions, err: tt.listerErr}
			picker := &fakePicker{name: tt.pickerName, err: tt.pickerErr}
			executor := &fakeExecutor{err: tt.executorErr}

			app := gomux.New(lister, picker, executor)
			err := app.Run(context.Background(), tt.sessionName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if lister.called != tt.wantListerCall {
				t.Errorf("lister.called = %v, want %v", lister.called, tt.wantListerCall)
			}

			if picker.called != tt.wantPickerCall {
				t.Errorf("picker.called = %v, want %v", picker.called, tt.wantPickerCall)
			}

			if executor.called != tt.wantExecutorCall {
				t.Errorf("executor.called = %v, want %v", executor.called, tt.wantExecutorCall)
			}

			if executor.called && executor.sessionName != tt.wantExecName {
				t.Errorf("executor.sessionName = %q, want %q", executor.sessionName, tt.wantExecName)
			}
		})
	}
}

// Discrete tests for error propagation and session passing

func TestAppRun_ListerErrorPropagation(t *testing.T) {
	t.Parallel()

	listerErr := fmt.Errorf("test error: %w", errTmuxServerNotRunning)
	lister := &fakeLister{err: listerErr}
	picker := &fakePicker{}
	executor := &fakeExecutor{}
	app := gomux.New(lister, picker, executor)

	err := app.Run(context.Background(), "")
	if err == nil {
		t.Fatal("Run() expected error, got nil")
	}

	if !errors.Is(err, listerErr) {
		t.Errorf("Run() error should wrap lister error, got %v", err)
	}

	if picker.called {
		t.Error("picker should not be called when lister fails")
	}

	if executor.called {
		t.Error("executor should not be called when lister fails")
	}
}

func TestAppRun_ExecutorErrorPropagation(t *testing.T) {
	t.Parallel()

	executorErr := fmt.Errorf("test error: %w", errTmuxBinaryNotFound)
	lister := &fakeLister{
		sessions: []gomux.Session{
			{Name: "dev", Windows: 1, Created: time.Now()},
		},
	}
	picker := &fakePicker{name: "dev"}
	executor := &fakeExecutor{err: executorErr}
	app := gomux.New(lister, picker, executor)

	err := app.Run(context.Background(), "")
	if err == nil {
		t.Fatal("Run() expected error, got nil")
	}

	if !errors.Is(err, executorErr) {
		t.Errorf("Run() error should wrap executor error, got %v", err)
	}
}

func TestAppRun_PickerReceivesSessions(t *testing.T) {
	t.Parallel()

	sessions := []gomux.Session{
		{Name: "dev", Windows: 2, Created: time.Now()},
		{Name: "prod", Windows: 5, Attached: true, Created: time.Now()},
	}

	lister := &fakeLister{sessions: sessions}
	picker := &fakePicker{name: "dev"}
	executor := &fakeExecutor{}

	app := gomux.New(lister, picker, executor)
	_ = app.Run(context.Background(), "")

	if len(picker.received) != len(sessions) {
		t.Fatalf("picker received %d sessions, want %d", len(picker.received), len(sessions))
	}

	for i, session := range sessions {
		if picker.received[i] != session {
			t.Errorf("picker.received[%d] = %+v, want %+v", i, picker.received[i], session)
		}
	}
}
