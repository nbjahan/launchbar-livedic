package launchbar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"time"
)

// ConfigValues represents a Config values
type ConfigValues map[string]interface{}

// Config provides permanent config utils for the action.
type Config struct {
	path string
	data map[string]interface{}
}

// NewConfig initializes an new Config object with the specified path and returns it.
func NewConfig(p string) *Config {
	return loadConfig(p)
}

// NewConfigDefaults initializes a new Config object with the specified path
// and default values and returns it.
func NewConfigDefaults(p string, defaults ConfigValues) *Config {
	config := loadConfig(p)
	for k, v := range defaults {
		if _, found := config.data[k]; !found {
			config.data[k] = v
		}
	}
	config.save()
	return config
}

// Delete removes the key from config file.
func (c *Config) Delete(keys ...string) {
	for _, key := range keys {
		delete(c.data, key)
	}
	c.save()
}

// Set sets the key, val and saves the config to the disk.
func (c *Config) Set(key string, val interface{}) {
	if !path.IsAbs(c.path) || path.Dir(path.Dir(c.path)) != os.ExpandEnv("$HOME/Library/Application Support/LaunchBar/Action Support") {
		panic(fmt.Sprintf("bad config path: %q", c.path))
	}

	c.data[key] = val
	c.save()
}

// Get gets the value from config for the key
func (c *Config) Get(key string) interface{} {
	return c.data[key]
}

// GetString gets the value from config for the key as string
func (c *Config) GetString(key string) string {
	if c.data[key] == nil {
		return ""
	}
	return fmt.Sprintf("%v", c.data[key])
}

// GetInt gets the value from config for the key as int64
func (c *Config) GetInt(key string) int64 {
	if c.data[key] == nil {
		return 0
	}
	i, ok := c.data[key].(float64)
	if !ok {
		return 0
	}
	return int64(i)
}

// GetFloat gets the value from config for the key as float64
func (c *Config) GetFloat(key string) float64 {
	if c.data[key] == nil {
		return 0
	}

	i, ok := c.data[key].(float64)
	if !ok {
		return 0
	}
	return i
}

// GetBool gets the value from config for the key as bool
func (c *Config) GetBool(key string) bool {
	if c.data[key] == nil {
		return false
	}

	b, ok := c.data[key].(bool)
	if !ok {
		return false
	}
	return b
}

// GetTimeDuration gets the value from config for the key as time.Duration
func (c *Config) GetTimeDuration(key string) time.Duration {
	if c.data[key] == nil {
		return time.Duration(0)
	}

	d, ok := c.data[key].(float64)
	if !ok {
		return 0
	}
	return time.Duration(d)
}

func loadConfig(p string) *Config {
	p = path.Join(p, "config.json")
	config := &Config{path: p, data: make(ConfigValues)}

	if data, err := ioutil.ReadFile(p); err == nil {
		json.Unmarshal(data, &config.data)
	}
	return config
}

func (c *Config) save() {
	if data, err := json.Marshal(&c.data); err == nil {
		ioutil.WriteFile(c.path, data, 0664)
	}
}
