package profile

import "testing"

func TestNew(t *testing.T) {
	p, err := New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "work" || p.Name != "Work" {
		t.Fatalf("unexpected profile: %#v", p)
	}
}
