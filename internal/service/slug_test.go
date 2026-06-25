package service

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Customer Satisfaction": "customer-satisfaction",
		"  Hello,  World!!  ":   "hello-world",
		"Café déjà vu":          "caf-d-j-vu",
		"":                      "form",
		"---":                   "form",
		"Already-Slugged-2024":  "already-slugged-2024",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRandSuffix(t *testing.T) {
	s := randSuffix(6)
	if len(s) != 6 {
		t.Fatalf("want length 6, got %d (%q)", len(s), s)
	}
	for _, r := range s {
		if !strings.ContainsRune(slugAlphabet, r) {
			t.Fatalf("char %q not in alphabet", r)
		}
	}
	if randSuffix(6) == randSuffix(6) {
		t.Error("expected distinct suffixes across calls")
	}
}
