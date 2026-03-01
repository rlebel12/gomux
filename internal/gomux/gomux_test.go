package gomux_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	err              error
	received         []gomux.Session
	receivedProfiles []gomux.Profile
	pickerSelection  gomux.Selection
	called           bool
}

func (f *fakePicker) Pick(_ context.Context, sessions []gomux.Session, profiles []gomux.Profile) (gomux.Selection, error) {
	f.called = true
	f.received = sessions
	f.receivedProfiles = profiles

	return f.pickerSelection, f.err
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

type fakeProfileLoader struct {
	profiles []gomux.Profile
	err      error
	called   bool
}

func (f *fakeProfileLoader) LoadProfiles(_ context.Context) ([]gomux.Profile, error) {
	f.called = true
	return f.profiles, f.err
}

type fakeSessionCreator struct {
	received gomux.Profile
	err      error
	called   bool
}

func (f *fakeSessionCreator) CreateSessionFromProfile(_ context.Context, profile gomux.Profile) error {
	f.called = true
	f.received = profile

	return f.err
}

// Parameterized tests for App.Run routing logic

func TestAppRun(t *testing.T) {
	t.Parallel()

	devProfile := gomux.Profile{Name: "dev", Windows: []gomux.Window{{Name: "editor"}}}

	tests := []struct {
		listerErr           error
		pickerErr           error
		executorErr         error
		listerSessions      []gomux.Session
		profileLoaderResult []gomux.Profile
		pickerSelection     gomux.Selection
		wantCreatorProfile  *gomux.Profile
		wantErr             bool
		wantListerCall      bool
		wantPickerCall      bool
		wantExecutorCall    bool
		wantCreatorCall     bool
		name                string
		sessionName         string
		wantExecName        string
	}{
		{
			name:             "direct_mode",
			sessionName:      "work",
			wantExecName:     "work",
			wantListerCall:   false,
			wantPickerCall:   false,
			wantExecutorCall: true,
			wantCreatorCall:  false,
		},
		{
			name:                "empty_state",
			sessionName:         "",
			listerSessions:      []gomux.Session{},
			profileLoaderResult: []gomux.Profile{},
			wantExecName:        gomux.DefaultSessionName,
			wantListerCall:      true,
			wantPickerCall:      false,
			wantExecutorCall:    true,
			wantCreatorCall:     false,
		},
		{
			name:        "tui_mode_single",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 1, Created: time.Now()},
			},
			profileLoaderResult: []gomux.Profile{},
			pickerSelection:     gomux.Selection{Name: "dev", IsProfile: false},
			wantExecName:        "dev",
			wantListerCall:      true,
			wantPickerCall:      true,
			wantExecutorCall:    true,
			wantCreatorCall:     false,
		},
		{
			name:        "tui_mode_multi",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 2, Created: time.Now()},
				{Name: "api", Windows: 1, Created: time.Now()},
			},
			profileLoaderResult: []gomux.Profile{},
			pickerSelection:     gomux.Selection{Name: "api", IsProfile: false},
			wantExecName:        "api",
			wantListerCall:      true,
			wantPickerCall:      true,
			wantExecutorCall:    true,
			wantCreatorCall:     false,
		},
		{
			name:        "user_quit",
			sessionName: "",
			listerSessions: []gomux.Session{
				{Name: "dev", Windows: 1, Created: time.Now()},
			},
			profileLoaderResult: []gomux.Profile{},
			pickerErr:           gomux.ErrNoSelection,
			wantListerCall:      true,
			wantPickerCall:      true,
			wantExecutorCall:    false,
			wantCreatorCall:     false,
		},
		{
			name:           "profile_selected_creates_session",
			sessionName:    "",
			listerSessions: []gomux.Session{{Name: "existing", Windows: 1, Created: time.Now()}},
			profileLoaderResult: []gomux.Profile{
				devProfile,
			},
			pickerSelection:    gomux.Selection{Name: "dev", IsProfile: true},
			wantExecName:       "dev",
			wantListerCall:     true,
			wantPickerCall:     true,
			wantExecutorCall:   true,
			wantCreatorCall:    true,
			wantCreatorProfile: &devProfile,
		},
		{
			name:           "session_selected_skips_creation",
			sessionName:    "",
			listerSessions: []gomux.Session{{Name: "existing", Windows: 1, Created: time.Now()}},
			profileLoaderResult: []gomux.Profile{
				devProfile,
			},
			pickerSelection:  gomux.Selection{Name: "existing", IsProfile: false},
			wantExecName:     "existing",
			wantListerCall:   true,
			wantPickerCall:   true,
			wantExecutorCall: true,
			wantCreatorCall:  false,
		},
		{
			name:                "no_profiles_file_empty_list",
			sessionName:         "",
			listerSessions:      []gomux.Session{{Name: "existing", Windows: 1, Created: time.Now()}},
			profileLoaderResult: []gomux.Profile{},
			pickerSelection:     gomux.Selection{Name: "existing", IsProfile: false},
			wantExecName:        "existing",
			wantListerCall:      true,
			wantPickerCall:      true,
			wantExecutorCall:    true,
			wantCreatorCall:     false,
		},
		{
			name:           "direct_mode_skips_profiles",
			sessionName:    "mysession",
			listerSessions: []gomux.Session{},
			profileLoaderResult: []gomux.Profile{
				devProfile,
			},
			wantExecName:     "mysession",
			wantListerCall:   false,
			wantPickerCall:   false,
			wantExecutorCall: true,
			wantCreatorCall:  false,
		},
		{
			name:           "no_sessions_but_profiles_shows_picker",
			sessionName:    "",
			listerSessions: []gomux.Session{},
			profileLoaderResult: []gomux.Profile{
				devProfile,
			},
			pickerSelection:    gomux.Selection{Name: "dev", IsProfile: true},
			wantExecName:       "dev",
			wantListerCall:     true,
			wantPickerCall:     true,
			wantExecutorCall:   true,
			wantCreatorCall:    true,
			wantCreatorProfile: &devProfile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lister := &fakeLister{sessions: tt.listerSessions, err: tt.listerErr}
			picker := &fakePicker{pickerSelection: tt.pickerSelection, err: tt.pickerErr}
			executor := &fakeExecutor{err: tt.executorErr}
			loader := &fakeProfileLoader{profiles: tt.profileLoaderResult}
			creator := &fakeSessionCreator{}

			app := gomux.New(lister, picker, executor, loader, creator)
			err := app.Run(t.Context(), tt.sessionName)

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

			if creator.called != tt.wantCreatorCall {
				t.Errorf("creator.called = %v, want %v", creator.called, tt.wantCreatorCall)
			}

			if tt.wantCreatorProfile != nil && !reflect.DeepEqual(creator.received, *tt.wantCreatorProfile) {
				t.Errorf("creator.received = %+v, want %+v", creator.received, *tt.wantCreatorProfile)
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
	loader := &fakeProfileLoader{}
	creator := &fakeSessionCreator{}
	app := gomux.New(lister, picker, executor, loader, creator)

	err := app.Run(t.Context(), "")
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
	picker := &fakePicker{pickerSelection: gomux.Selection{Name: "dev", IsProfile: false}}
	executor := &fakeExecutor{err: executorErr}
	loader := &fakeProfileLoader{}
	creator := &fakeSessionCreator{}
	app := gomux.New(lister, picker, executor, loader, creator)

	err := app.Run(t.Context(), "")
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
	profiles := []gomux.Profile{
		{Name: "staging", Windows: []gomux.Window{{Name: "editor"}}},
	}

	lister := &fakeLister{sessions: sessions}
	picker := &fakePicker{pickerSelection: gomux.Selection{Name: "dev", IsProfile: false}}
	executor := &fakeExecutor{}
	loader := &fakeProfileLoader{profiles: profiles}
	creator := &fakeSessionCreator{}

	app := gomux.New(lister, picker, executor, loader, creator)
	_ = app.Run(t.Context(), "")

	wantProfiles := profiles

	if len(picker.received) != len(sessions) {
		t.Fatalf("picker received %d sessions, want %d", len(picker.received), len(sessions))
	}

	for i, session := range sessions {
		if picker.received[i] != session {
			t.Errorf("picker.received[%d] = %+v, want %+v", i, picker.received[i], session)
		}
	}

	if len(picker.receivedProfiles) != len(wantProfiles) {
		t.Fatalf("picker received %d profiles, want %d", len(picker.receivedProfiles), len(wantProfiles))
	}

	for i, profile := range wantProfiles {
		if picker.receivedProfiles[i].Name != profile.Name {
			t.Errorf("picker.receivedProfiles[%d].Name = %q, want %q", i, picker.receivedProfiles[i].Name, profile.Name)
		}
	}
}

func TestAppRun_ProfileLoaderErrorPropagation(t *testing.T) {
	t.Parallel()

	loaderErr := errors.New("failed to read profiles file")
	lister := &fakeLister{
		sessions: []gomux.Session{
			{Name: "dev", Windows: 1, Created: time.Now()},
		},
	}
	picker := &fakePicker{}
	executor := &fakeExecutor{}
	loader := &fakeProfileLoader{err: loaderErr}
	creator := &fakeSessionCreator{}
	app := gomux.New(lister, picker, executor, loader, creator)

	err := app.Run(t.Context(), "")
	if err == nil {
		t.Fatal("Run() expected error, got nil")
	}

	if !errors.Is(err, loaderErr) {
		t.Errorf("Run() error should wrap loader error, got %v", err)
	}

	if picker.called {
		t.Error("picker should not be called when loader fails")
	}

	if creator.called {
		t.Error("creator should not be called when loader fails")
	}
}

func TestAppRun_SessionCreatorErrorPropagation(t *testing.T) {
	t.Parallel()

	creatorErr := errors.New("failed to create tmux session")
	lister := &fakeLister{
		sessions: []gomux.Session{
			{Name: "existing", Windows: 1, Created: time.Now()},
		},
	}
	devProfile := gomux.Profile{Name: "dev", Windows: []gomux.Window{{Name: "editor"}}}
	picker := &fakePicker{pickerSelection: gomux.Selection{Name: "dev", IsProfile: true}}
	executor := &fakeExecutor{}
	loader := &fakeProfileLoader{profiles: []gomux.Profile{devProfile}}
	creator := &fakeSessionCreator{err: creatorErr}
	app := gomux.New(lister, picker, executor, loader, creator)

	err := app.Run(t.Context(), "")
	if err == nil {
		t.Fatal("Run() expected error, got nil")
	}

	if !errors.Is(err, creatorErr) {
		t.Errorf("Run() error should wrap creator error, got %v", err)
	}

	if executor.called {
		t.Error("executor should not be called when creator fails")
	}
}

func TestAppRun_PickerReceivesFilteredProfiles(t *testing.T) {
	t.Parallel()

	// Sessions "dev" and "api" already exist; profile "dev" should be excluded
	// because a session with the same name already exists.
	sessions := []gomux.Session{
		{Name: "dev", Windows: 1, Created: time.Now()},
		{Name: "api", Windows: 2, Created: time.Now()},
	}
	profiles := []gomux.Profile{
		{Name: "dev", Windows: []gomux.Window{{Name: "editor"}}},  // should be filtered out
		{Name: "staging", Windows: []gomux.Window{{Name: "web"}}}, // should pass through
	}

	lister := &fakeLister{sessions: sessions}
	picker := &fakePicker{pickerSelection: gomux.Selection{Name: "api", IsProfile: false}}
	executor := &fakeExecutor{}
	loader := &fakeProfileLoader{profiles: profiles}
	creator := &fakeSessionCreator{}

	app := gomux.New(lister, picker, executor, loader, creator)
	_ = app.Run(t.Context(), "")

	// Only "staging" should be passed to the picker; "dev" is filtered
	if len(picker.receivedProfiles) != 1 {
		t.Fatalf("picker received %d profiles, want 1", len(picker.receivedProfiles))
	}

	if picker.receivedProfiles[0].Name != "staging" {
		t.Errorf("picker.receivedProfiles[0].Name = %q, want %q", picker.receivedProfiles[0].Name, "staging")
	}
}
