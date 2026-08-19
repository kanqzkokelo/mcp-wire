package config

import "os"

type EnvConfig struct {
	StateDir string
	Token    string
	LogLevel string
}

func LoadEnvConfig() EnvConfig {
	return EnvConfig{
		StateDir: os.Getenv("MCP_WIRE_PATH"),
		Token:    os.Getenv("MCP_WIRE_TOKEN"),
		LogLevel: os.Getenv("MCP_WIRE_LOG_LEVEL"),
	}
}
