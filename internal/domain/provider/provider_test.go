package provider

import "testing"

func TestNew(t *testing.T) {
	p, err := New("cpa", "CPA", "http://127.0.0.1:8317/v1/", "credential:cpa")
	if err != nil {
		t.Fatal(err)
	}
	if p.BaseURL != "http://127.0.0.1:8317/v1" || p.Protocol != ProtocolResponses {
		t.Fatalf("unexpected provider: %#v", p)
	}
	if _, err := New("bad", "Bad", "not-a-url", ""); err != ErrInvalidBaseURL {
		t.Fatalf("expected invalid URL, got %v", err)
	}
}
