package config

import (
	"testing"
)

func TestParseCommaSeparatedEnv(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	if got := parseCommaSeparatedEnv("CORS_ALLOWED_ORIGINS"); got != nil {
		t.Fatalf("expected nil when env is unset, got %#v", got)
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", " https://planet.zhelearn.com ")
	got := parseCommaSeparatedEnv("CORS_ALLOWED_ORIGINS")
	if len(got) != 1 || got[0] != "https://planet.zhelearn.com" {
		t.Fatalf("unexpected single value parse result: %#v", got)
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", " https://planet.zhelearn.com,https://admin.zhelearn.com , https://planet.zhelearn.com ,,")
	got = parseCommaSeparatedEnv("CORS_ALLOWED_ORIGINS")
	if len(got) != 2 || got[0] != "https://planet.zhelearn.com" || got[1] != "https://admin.zhelearn.com" {
		t.Fatalf("unexpected multi value parse result: %#v", got)
	}
}
