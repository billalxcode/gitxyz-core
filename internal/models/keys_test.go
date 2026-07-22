package models

import "testing"

func TestParseScopes(t *testing.T) {
	got := ParseScopes("repo:read,repo:write, admin:*")
	want := []string{"repo:read", "repo:write", "admin:*"}
	if len(got) != len(want) {
		t.Fatalf("expected %d scopes, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("scope %d: want %q got %q", i, want[i], got[i])
		}
	}
}
