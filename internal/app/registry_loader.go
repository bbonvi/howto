package app

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/loader"
	"github.com/yourusername/howto/internal/parser"
	"github.com/yourusername/howto/internal/registry"
)

// RegistryLoader exposes a cached view of the playbook registry.
type RegistryLoader interface {
	Load() (registry.Registry, error)
}

// CachedRegistryLoader caches the playbook registry and reloads when source files change.
type CachedRegistryLoader struct {
	mu         sync.Mutex
	globalDir  string
	projectDir string

	cached    registry.Registry
	signature string
}

// NewCachedRegistryLoader creates a new CachedRegistryLoader rooted at the provided directories.
func NewCachedRegistryLoader(globalDir, projectDir string) *CachedRegistryLoader {
	return &CachedRegistryLoader{
		globalDir:  globalDir,
		projectDir: projectDir,
	}
}

// LoadRegistry builds the registry from disk without caching.
func LoadRegistry(globalDir, projectDir string) (registry.Registry, error) {
	globalDocs, err := loader.LoadGlobalDocs(globalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load global docs: %w", err)
	}

	projectDocs, err := loader.LoadProjectDocs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load project docs: %w", err)
	}

	projectConfig, err := config.LoadProjectConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	reg := registry.BuildRegistry(globalDocs, projectDocs, projectConfig)
	return reg, nil
}

// Load returns the cached registry, reloading from disk if the source documents changed.
func (c *CachedRegistryLoader) Load() (registry.Registry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentSignature, err := computeSignature(c.globalDir, c.projectDir)
	if err != nil {
		return nil, err
	}

	if c.cached != nil && c.signature == currentSignature {
		return cloneRegistry(c.cached), nil
	}

	reg, err := LoadRegistry(c.globalDir, c.projectDir)
	if err != nil {
		return nil, err
	}

	c.cached = reg
	c.signature = currentSignature

	return cloneRegistry(c.cached), nil
}

func cloneRegistry(src registry.Registry) registry.Registry {
	if src == nil {
		return nil
	}

	dest := make(registry.Registry, len(src))
	for k, v := range src {
		dest[k] = v
	}
	return dest
}

func computeSignature(dirs ...string) (string, error) {
	hasher := sha256.New()

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		info, err := os.Stat(dir)
		if errors.Is(err, os.ErrNotExist) {
			hasher.Write([]byte(dir))
			hasher.Write([]byte(":missing;"))
			continue
		} else if err != nil {
			return "", fmt.Errorf("failed to stat directory %s: %w", dir, err)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("path %s is not a directory", dir)
		}

		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				relPath = path
			}

			hasher.Write([]byte(relPath))
			hasher.Write([]byte{':'})
			hasher.Write([]byte(fmt.Sprintf("%d", info.ModTime().UnixNano())))
			hasher.Write([]byte{':'})
			hasher.Write([]byte(fmt.Sprintf("%d", info.Size())))
			hasher.Write([]byte{';'})
			return nil
		})

		if err != nil {
			return "", fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// DocumentsToList converts a registry into a sorted slice of documents.
func DocumentsToList(reg registry.Registry) []parser.Document {
	names := reg.List()
	docs := make([]parser.Document, 0, len(names))
	for _, name := range names {
		docs = append(docs, reg[name])
	}
	return docs
}

// SortedKeys returns registry keys sorted alphabetically.
func SortedKeys(reg registry.Registry) []string {
	keys := make([]string, 0, len(reg))
	for k := range reg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
