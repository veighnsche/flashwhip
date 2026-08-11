package config

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("FLASHWHIP_MODEL")
	os.Unsetenv("FLASHWHIP_URL")
	os.Unsetenv("FLASHWHIP_API_KEY")

	cfg := LoadConfig()
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("cfg.BaseURL = %q, want default %q", cfg.BaseURL, DefaultBaseURL)
	}

	if cfg.ModelName != DefaultModelName {
		t.Errorf("cfg.ModelName = %q, want default %q", cfg.ModelName, DefaultModelName)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("FLASHWHIP_MODEL", "qwen2.5-coder:7b")
	t.Setenv("FLASHWHIP_URL", "http://localhost:11434/v1")
	t.Setenv("FLASHWHIP_API_KEY", "test-key")

	cfg := LoadConfig()
	if cfg.ModelName != "qwen2.5-coder:7b" {
		t.Errorf("cfg.ModelName = %q, want 'qwen2.5-coder:7b'", cfg.ModelName)
	}

	if cfg.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("cfg.BaseURL = %q, want 'http://localhost:11434/v1'", cfg.BaseURL)
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("cfg.APIKey = %q, want 'test-key'", cfg.APIKey)
	}
}

func TestBuildProjectContextIncludesRuntime(t *testing.T) {
	context := BuildProjectContext(t.TempDir())
	if !strings.Contains(context, runtime.GOOS+"/"+runtime.GOARCH) || !strings.Contains(context, runtime.Version()) {
		t.Fatalf("BuildProjectContext() missing runtime: %q", context)
	}
}

func TestSetAndGetProjectRoot(t *testing.T) {
	customDir := "/tmp/custom-project-root"
	SetProjectRoot(customDir)

	got := GetProjectRoot()
	if got != customDir {
		t.Errorf("GetProjectRoot() = %q, want %q", got, customDir)
	}
}
