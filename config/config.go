package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load(dotEnvFile, configYamlFile string) Config {
	envMap := loadDotEnv(dotEnvFile)
	yamlMap := loadYAML(configYamlFile)

	dbURL := getString("DATABASE_URL", envMap, yamlMap, "") // required
	port := getString("PORT", envMap, yamlMap, "8080")

	if dbURL == "" {
		panic("missing required config: DATABASE_URL")
	}

	// Валидация порта uint 16 bit
	if port != "" {
		if n, err := strconv.ParseUint(port, 10, 16); err != nil || n == 0 {
			panic(fmt.Errorf("invalid PORT value: %q", port))
		}
	}

	return Config{
		DatabaseURL: dbURL,
		Port:        port,
	}
}

func getString(key string, envMap, yamlMap map[string]string, defaultValue string) string {
	// .env
	if v, ok := envMap[key]; ok && v != "" {
		return v
	}

	// config.yaml
	if v, ok := yamlMap[key]; ok && v != "" {
		return v
	}
	lc := strings.ToLower(key)
	alts := []string{lc, toSnake(lc), toKebab(lc)}
	for _, k := range alts {
		if v, ok := yamlMap[k]; ok && v != "" {
			return v
		}
	}

	// environment
	if v := os.Getenv(key); v != "" {
		return v
	}

	// default
	return defaultValue
}

func loadDotEnv(path string) map[string]string {
	m := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		val = trimQuotes(val)
		if key != "" {
			m[key] = val
		}
	}
	return m
}

func loadYAML(path string) map[string]string {
	m := make(map[string]string)

	var f *os.File
	f, _ = os.Open(path)

	if f == nil {
		return m
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		val = strings.TrimLeft(val, " ")
		val = trimQuotes(val)
		if key != "" {
			m[key] = val
		}
	}
	return m
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func toSnake(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func toKebab(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
