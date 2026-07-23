package config

import (
	"testing"
	"time"
)

func TestLoadBusinessRequestTimeout(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "default", want: 2 * time.Second},
		{name: "milliseconds", value: "500ms", want: 500 * time.Millisecond},
		{name: "seconds", value: "3s", want: 3 * time.Second},
		{name: "invalid", value: "soon", wantErr: true},
		{name: "zero", value: "0s", wantErr: true},
		{name: "negative", value: "-1s", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setRequiredConfigEnv(t)
			t.Setenv("BUSINESS_REQUEST_TIMEOUT", test.value)

			cfg, err := Load()
			if test.wantErr {
				if err == nil {
					t.Fatal("expected invalid business request timeout to fail")
				}
				return
			}
			if err != nil {
				t.Fatalf("load config: %v", err)
			}
			if cfg.BusinessRequestTimeout != test.want {
				t.Fatalf("unexpected business request timeout: got %s, want %s", cfg.BusinessRequestTimeout, test.want)
			}
		})
	}
}

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

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"DB_HOST": "localhost", "DB_PORT": "5432", "DB_USER": "postgres",
		"DB_PASSWORD": "postgres", "DB_NAME": "planet_test", "DB_SSLMODE": "disable",
		"DB_TIMEZONE": "Asia/Shanghai", "ADMIN_BEARER_TOKEN": "test-token",
		"SWAGGER_PASSWORD": "test-password",
	} {
		t.Setenv(key, value)
	}
}
