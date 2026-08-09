package config

import (
	"os"
	"strings"
)

// Config holds runtime configuration for the ADK 2.0 CLI application.
type Config struct {
	BaseURL           string
	ModelName         string
	APIKey            string
	SystemInstruction string
	// ProjectRoot is the working directory the agent operates in.
	// Defaults to os.Getwd() at startup; overridable via --cwd flag.
	ProjectRoot string
}

const (
	DefaultBaseURL     = "https://ollama.dimensionlab.net/v1"
	DefaultModelName   = "hf.co/gbuzhf/KAT-Coder-V2.5-Dev-MTP-GGUF:UD-Q4_K_XL"
	DefaultSystem      = "You are Flashwhip, an intelligent AI coding and terminal assistant powered by Google ADK 2.0. Provide concise, helpful, and technically accurate responses."
	DefaultUserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	MaxContentLength   = 4000
	MaxSearchResults   = 6
)

// LoadConfig creates a Config instance, pulling from environment variables with fallback defaults.
func LoadConfig() *Config {
	baseURL := os.Getenv("FLASHWHIP_URL")
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
	}
	if baseURL == "" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if !strings.HasSuffix(baseURL, "/v1") && !strings.HasSuffix(baseURL, "/v1/") {
		baseURL = strings.TrimRight(baseURL, "/") + "/v1"
	}

	modelName := os.Getenv("FLASHWHIP_MODEL")
	if modelName == "" {
		modelName = os.Getenv("OLLAMA_MODEL")
	}
	if modelName == "" {
		modelName = os.Getenv("OPENAI_MODEL")
	}
	if modelName == "" {
		modelName = DefaultModelName
	}

	apiKey := os.Getenv("FLASHWHIP_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OLLAMA_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		apiKey = "ollama"
	}

	cwd, _ := os.Getwd()

	return &Config{
		BaseURL:           baseURL,
		ModelName:         modelName,
		APIKey:            apiKey,
		SystemInstruction: DefaultSystem,
		ProjectRoot:       cwd,
	}
}
