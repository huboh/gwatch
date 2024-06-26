// Package config provides functionality loading & writing our app config.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/huboh/gwatch/internal/pkg/utils"
	"gopkg.in/yaml.v3"
)

var (
	// rootDir is the current working directory
	rootDir = utils.Must(os.Getwd())

	// configName is our configuration file name
	configName = "gwatch.yml"

	// ConfigPath is path to our configuration file
	ConfigPath = filepath.Join(rootDir, configName)

	// defaultExts defines the default file extensions to watch for changes.
	defaultExts = []string{"go", "tmp", "tmpl", "html"}

	// defaultPaths defines the default paths to watch for changes.
	defaultPaths = []string{rootDir}

	// defaultExclude defines the default directories to exclude from watching.
	defaultExclude = []string{".git", "bin", "vendor", "testdata"}

	// defaultRecursive defines whether to watch directories listed in `defaultPaths` recursively.
	defaultRecursive = true

	// defaultBuildCmd is the command used to build the project.
	defaultBuildCmd = fmt.Sprintf("go build -o %s %s", defaultBinPath, rootDir)

	// defaultDelayMs is the watcher delay in between events
	defaultDelay = time.Millisecond * 100

	// defaultLogPrefix is the prefix added to runner stderr/stdout output
	defaultLogPrefix = filepath.Base(rootDir)
)

// Config represents the app's
type Config struct {
	// watcher config
	Root      string        `yaml:"root"`
	Exts      []string      `yaml:"exts,flow"`
	Paths     []string      `yaml:"paths,flow"`
	Exclude   []string      `yaml:"exclude,flow"`
	Delay     time.Duration `yaml:"delay"`
	Recursive bool          `yaml:"recursive"`

	// runner config
	LogPrefix string      `yaml:"log_prefix"`
	Run       RunConfig   `yaml:"run"`
	Build     BuildConfig `yaml:"build"`
}

// Run represents the run configuration for the runner.
type RunConfig struct {
	// Bin is the binary to be executed.
	Bin string `yaml:"bin"`

	// Args are the arguments to be passed to the binary.
	Args []string `yaml:"args,flow"`
}

// Build represents the build configuration for the runner.
type BuildConfig struct {
	// Cmd is the build command to be executed.
	Cmd string `yaml:"cmd"`
}

// New reads the config file in the root directory and returns it.
// If the config file does not exist, it creates a new config file with the defaults and returns it.
//
// The config file is expected to be in YAML format.
//
// It returns a pointer to a Config and an error. If successful, the error is nil.
func New() (*Config, error) {
	var (
		config  = Default()
		loadErr error
	)

	defer func() {
		if val := recover(); val != nil {
			if err, ok := val.(error); ok {
				loadErr = fmt.Errorf("error loading config file: %w", err)
			}
		}
	}()

	// if config file don't exists create new one and write our defaults to it, then return it.
	if _, err := os.Stat(ConfigPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err := createAndWriteConfigFile(ConfigPath, *config); err != nil {
			return nil, err
		}

		return config, nil
	}

	// read config file and merge it with our defaults, then return it.
	byts, err := os.ReadFile(ConfigPath)

	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(byts, config); err != nil {
		return nil, err
	}

	return config, loadErr
}

// Default returns a pointer to a new Config initialized with the default values.
func Default() *Config {
	return &Config{
		Root:      rootDir,
		Exts:      defaultExts,
		Paths:     defaultPaths,
		Exclude:   defaultExclude,
		Delay:     defaultDelay,
		Recursive: defaultRecursive,
		LogPrefix: defaultLogPrefix,

		Run: RunConfig{
			Bin:  defaultBinPath,
			Args: []string{},
		},

		Build: BuildConfig{
			Cmd: defaultBuildCmd,
		},
	}
}

// createAndWriteConfigFile creates a new config file at the specified path and writes the provided config to it.
//
// Returns an error if the file creation or writing process fails.
func createAndWriteConfigFile(path string, config Config) error {
	file, err := os.Create(path)

	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}

	defer file.Close()

	// write config to file
	// yaml.
	if err := yaml.NewEncoder(file).Encode(config); err != nil {

		// delete the new config file incase of write error
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("error deleting config file after write failure: %w", err)
		}

		return fmt.Errorf("error writing new config file: %w", err)
	}

	return nil
}

// Reload reloads the configuration from the config file, updating the current Config instance.
//
// It reads the config file from the root directory and updates the fields of the current Config instance.
//
// Returns an error if there was an issue reading or parsing the config file.
func (c *Config) Reload() error {
	config, err := New()
	if err != nil {
		return err
	}

	*c = *config
	return nil
}
