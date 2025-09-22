package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalBackend implements Backend for local filesystem storage
type LocalBackend struct {
	config *Config
	baseDir string
}

// NewLocalBackend creates a new local filesystem backend
func NewLocalBackend(config *Config) (*LocalBackend, error) {
	baseDir := "/"
	if config.Path != "" {
		baseDir = config.Path
	}

	return &LocalBackend{
		config: config,
		baseDir: baseDir,
	}, nil
}

// localWriter wraps an os.File to implement Writer interface
type localWriter struct {
	file *os.File
	path string
}

func (w *localWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

func (w *localWriter) Close() error {
	return w.file.Close()
}

func (w *localWriter) URL() string {
	return w.path
}

// localReader wraps an os.File to implement Reader interface
type localReader struct {
	file *os.File
	path string
}

func (r *localReader) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

func (r *localReader) Close() error {
	return r.file.Close()
}

func (r *localReader) URL() string {
	return r.path
}

// NewWriter creates a new writer for local filesystem
func (b *LocalBackend) NewWriter(ctx context.Context, path string) (Writer, error) {
	fullPath := filepath.Join(b.baseDir, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0750); err != nil { // #nosec G301
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(fullPath) // #nosec G304 - path is joined with controlled base directory
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}

	return &localWriter{
		file: file,
		path: fullPath,
	}, nil
}

// NewReader creates a new reader for local filesystem
func (b *LocalBackend) NewReader(ctx context.Context, path string) (Reader, error) {
	fullPath := filepath.Join(b.baseDir, path)

	file, err := os.Open(fullPath) // #nosec G304 - path is joined with controlled base directory
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}

	return &localReader{
		file: file,
		path: fullPath,
	}, nil
}

// Exists checks if a file exists
func (b *LocalBackend) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(b.baseDir, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// List lists files with the given prefix
func (b *LocalBackend) List(ctx context.Context, prefix string) ([]string, error) {
	fullPrefix := filepath.Join(b.baseDir, prefix)
	dir := filepath.Dir(fullPrefix)
	base := filepath.Base(fullPrefix)

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(filepath.Base(path), base) {
			relPath, _ := filepath.Rel(b.baseDir, path)
			files = append(files, relPath)
		}
		return nil
	})

	return files, err
}

// Delete removes a file
func (b *LocalBackend) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(b.baseDir, path)
	return os.Remove(fullPath)
}

// Provider returns the provider type
func (b *LocalBackend) Provider() Provider {
	return ProviderLocal
}