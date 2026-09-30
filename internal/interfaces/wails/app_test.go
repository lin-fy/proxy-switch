package wails

import (
	"context"
	"errors"
	"testing"
)

type fakeAutostart struct {
	enabled bool
	err     error
}

func (f *fakeAutostart) Enable() error {
	if f.err == nil {
		f.enabled = true
	}
	return f.err
}

func (f *fakeAutostart) Disable() error {
	if f.err == nil {
		f.enabled = false
	}
	return f.err
}

func (f *fakeAutostart) IsEnabled() (bool, error) { return f.enabled, f.err }

func TestAutostartFacade(t *testing.T) {
	app := &App{}
	manager := &fakeAutostart{}
	app.autostart = manager

	if err := app.SetAutostart(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	enabled, err := app.AutostartEnabled(context.Background())
	if err != nil || !enabled {
		t.Fatalf("enabled = %v, err = %v", enabled, err)
	}

	if err := app.SetAutostart(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	enabled, err = app.AutostartEnabled(context.Background())
	if err != nil || enabled {
		t.Fatalf("enabled = %v, err = %v", enabled, err)
	}
}

func TestAutostartFacadeRequiresManager(t *testing.T) {
	app := &App{}
	if _, err := app.AutostartEnabled(context.Background()); !errors.Is(err, errAutostartUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}
