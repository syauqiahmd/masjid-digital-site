package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("# comment\n\nDOTENV_A=from file\nexport DOTENV_B=\"quoted value\"\nDOTENV_C='x'\nDOTENV_SET=from file\n"), 0o600)
	t.Setenv("DOTENV_SET", "from shell")
	for _, k := range []string{"DOTENV_A", "DOTENV_B", "DOTENV_C"} {
		t.Setenv(k, "") // registers cleanup
		os.Unsetenv(k)
	}

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}
	want := map[string]string{"DOTENV_A": "from file", "DOTENV_B": "quoted value", "DOTENV_C": "x", "DOTENV_SET": "from shell"}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), "nope")); err != nil {
		t.Errorf("loadDotEnv = %v, want nil", err)
	}
}

func TestLoadDotEnvBadLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("JUST_A_WORD\n"), 0o600)
	if err := loadDotEnv(path); err == nil {
		t.Error("loadDotEnv = nil, want an error")
	}
}
