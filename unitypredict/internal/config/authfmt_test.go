package config

import "testing"

// The prod UPT key contains { [ & * ) — the exact bytes that bricked the
// previous generated CLI when the unresolved-placeholder guard ran AFTER
// substitution. Header must survive intact.
func TestApplyAuthFormat_KeyContainingBraces(t *testing.T) {
	key := "abc{def[gh&ij*kl)mn"
	got := applyAuthFormat("Bearer APIKEY@{token}", map[string]string{"token": key})
	want := "Bearer APIKEY@" + key
	if got != want {
		t.Fatalf("header blanked or mangled\n got: %q\nwant: %q", got, want)
	}
}

func TestApplyAuthFormat_GenuinelyUnresolvedStillBlanks(t *testing.T) {
	if got := applyAuthFormat("Bearer {nope}", map[string]string{"token": "x"}); got != "" {
		t.Fatalf("expected blank for unresolved placeholder, got %q", got)
	}
}
