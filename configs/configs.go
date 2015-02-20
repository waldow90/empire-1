package configs

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"sync"

	"github.com/remind101/empire/apps"
)

// Version represents a unique identifier for a Config version.
type Version string

// Config represents a collection of environment variables.
type Config struct {
	Version Version   `json:"version"`
	App     *apps.App `json:"app"`
	Vars    Vars      `json:"vars"`
}

// NewConfig initializes a new config based on the old config, with the new
// variables provided.
func NewConfig(old *Config, vars Vars) *Config {
	v := mergeVars(old.Vars, vars)

	return &Config{
		Version: Version(hash(v)),
		App:     old.App,
		Vars:    v,
	}
}

// Variable represents the name of an environment variable.
type Variable string

// Vars represents a variable -> value mapping.
type Vars map[Variable]string

// Repository represents an interface for retrieving and storing Config's.
type Repository interface {
	// Head returns the current Config for the app.
	Head(apps.Name) (*Config, error)

	// Version returns the specific version of a Config for an app.
	Version(apps.Name, Version) (*Config, error)

	// Store stores the Config for the app.
	Push(*Config) (*Config, error)
}

func NewRepository() Repository {
	return newRepository()
}

// repository is an in memory implementation of the Repository.
type repository struct {
	sync.RWMutex
	s map[apps.Name][]*Config
	h map[apps.Name]*Config
}

func newRepository() *repository {
	return &repository{
		s: make(map[apps.Name][]*Config),
		h: make(map[apps.Name]*Config),
	}
}

// Head implements Repository Head.
func (r *repository) Head(appName apps.Name) (*Config, error) {
	r.RLock()
	defer r.RUnlock()

	if r.h[appName] == nil {
		return nil, nil
	}

	return r.h[appName], nil
}

// Version implements Repository Version.
func (r *repository) Version(appName apps.Name, version Version) (*Config, error) {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.s[appName] {
		if c.Version == version {
			return c, nil
		}
	}

	return nil, nil
}

// Push implements Repository Push.
func (r *repository) Push(config *Config) (*Config, error) {
	r.Lock()
	defer r.Unlock()

	n := config.App.Name

	r.s[n] = append(r.s[n], config)
	r.h[n] = config

	return config, nil
}

// mergeVars copies all of the vars from a, and merges b into them, returning a
// new Vars.
func mergeVars(old, new Vars) Vars {
	vars := make(Vars)

	for n, v := range old {
		vars[n] = v
	}
	for n, v := range new {
		if v != "" {
			vars[n] = v
		} else {
			delete(vars, n)
		}
	}

	return vars
}

// hash creates a sha1 hash of a set of Vars.
func hash(vars Vars) string {
	s := make(sort.StringSlice, len(vars))

	for n := range vars {
		s = append(s, string(n))
	}

	sort.Sort(s)

	v := ""

	for _, n := range s {
		v = v + n + "=" + vars[Variable(n)]
	}

	return fmt.Sprintf("%x", sha1.Sum([]byte(v)))
}
