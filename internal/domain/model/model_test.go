package model

import "testing"

func TestNewUsesIDAsDefaultName(t *testing.T) {
	m, err := New("cpa", "gpt-5", "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "gpt-5" || !m.Enabled {
		t.Fatalf("unexpected model: %#v", m)
	}
}
