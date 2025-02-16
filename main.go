package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jpappel/penny/api"
	"github.com/jpappel/penny/data"
	"github.com/jpappel/penny/text"
	"golang.org/x/oauth2"
)

type Config struct {
	Host           string   `json:"hostname"`
	Port           int      `json:"port"`
	BaseUrl        string   `json:"base_url"`
	RenderMD       bool     `json:"render_markdown"`
	Providers      []string `json:"providers"`
	EnvFilename    string   `json:"env_file"`
	EnabledFilters []string `json:"filters"`
	filters        []text.Filterer
	oauthConfigs   map[string]oauth2.Config
}

// Set env vars to values in a file
func readEnvfile(filename string) error {
	envFile, err := os.Open(filename)
	if err != nil {
		slog.Warn(fmt.Sprint("Unable to open environment file", filename))
		return err
	}
	defer envFile.Close()

	buf, err := io.ReadAll(envFile)
	if err != nil {
		slog.Error("Error occured while reading enviornment file")
		return err
	}

	if len(buf) > 2<<16 {
		slog.Warn("Environment file is larger than expected")
	}

	for _, line := range strings.Split(string(buf), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.TrimSpace(parts[0])
		if left[0] == '#' {
			continue
		}

		os.Setenv(left, parts[1])
	}

	return nil
}

// Parse a config file into a struct
func parseConfig(filename string) Config {
	cfgFile, err := os.Open(filename)
	if err != nil {
		slog.Error("Unable to open config file")
		panic(err)
	}

	buf, err := io.ReadAll(cfgFile)
	if err != nil {
		slog.Error("Error occured while reading config file")
		panic(err)
	}
	cfgFile.Close()

	cfg := Config{}
	if err := json.Unmarshal(buf, &cfg); err != nil {
		slog.Error("Unable to parse config file")
		panic(err)
	}

	if cfg.EnvFilename == "" {
		cfg.EnvFilename = ".env"
	}

	if err = readEnvfile(cfg.EnvFilename); err != nil {
		panic(err)
	}

	keys := []string{"GOOGLE_SECRET"}
	cfg.oauthConfigs = make(map[string]oauth2.Config)

	for _, key := range keys {
		val, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		// TODO: parse env vars
		cfg.oauthConfigs[val] = oauth2.Config{}
	}

	for _, filterName := range cfg.EnabledFilters {
		filter, ok := text.AvailableFilters[filterName]
		if !ok {
			slog.Error("Invalid Filter", slog.String("filterName", filterName))
			panic(fmt.Sprint("No filter:", filterName))
		}

		cfg.filters = append(cfg.filters, filter)
	}

	return cfg
}

func main() {
	// TODO: setup config loading hierarchy
	config := parseConfig("config.json")
	if config.Port <= 0 {
		config.Port = 8080
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	mux := api.NewMux(config.BaseUrl)

	data.New("file:data.sqlite3")

	slog.Info(fmt.Sprintf("Starting Penny on %s", addr))
	slog.Info(http.ListenAndServe(addr, mux).Error())
}
