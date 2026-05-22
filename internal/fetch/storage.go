package fetch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store is the minimal storage interface the Cloud Run job needs. It
// abstracts gs:// vs file:// so the same code path runs locally and on
// Cloud Run. Object paths are scheme-stripped (no gs:// or file://
// prefix), relative to the store root.
type Store interface {
	// Root returns the URI of the root location (e.g. "gs://bucket" or
	// "file:///abs/path"). Used to build full URIs for logging.
	Root() string

	// Write stores object contents at objectPath relative to the root.
	Write(ctx context.Context, objectPath string, data []byte, contentType string) error

	// Read returns the bytes at objectPath, or an error wrapping
	// ErrObjectNotFound if missing.
	Read(ctx context.Context, objectPath string) ([]byte, error)

	// List returns object paths under prefix in lexicographic order.
	List(ctx context.Context, prefix string) ([]string, error)

	// Close releases backend resources.
	Close() error
}

// ErrObjectNotFound is wrapped by Store.Read when an object is missing.
var ErrObjectNotFound = errors.New("object not found")

// NewStoreFromURI returns a Store backed by GCS or local disk depending
// on the URI scheme.
//
//   - gs://bucket          → GCS (production)
//   - file:///abs/path     → local (dev / tests)
//   - /abs/path            → local (dev / tests)
//
// The gs:// backend is intentionally a stub in the template. Wire it up
// against cloud.google.com/go/storage before the first Cloud Run
// deployment: add the dependency to go.mod, replace gcsStore below with
// a real implementation, and run `go mod tidy`.
func NewStoreFromURI(_ context.Context, uri string) (Store, error) {
	switch {
	case strings.HasPrefix(uri, "gs://"):
		return nil, fmt.Errorf("storage: gs:// backend not wired up yet — see internal/fetch/storage.go")
	case strings.HasPrefix(uri, "file://"):
		return newLocalStore(strings.TrimPrefix(uri, "file://"))
	case strings.HasPrefix(uri, "/"):
		return newLocalStore(uri)
	default:
		return nil, fmt.Errorf("storage: unsupported URI %q (want gs://, file://, or absolute path)", uri)
	}
}

type localStore struct {
	root string
}

func newLocalStore(root string) (*localStore, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("storage: resolve %q: %w", root, err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("storage: mkdir %q: %w", abs, err)
	}
	return &localStore{root: abs}, nil
}

func (s *localStore) Root() string { return "file://" + s.root }

func (s *localStore) Write(_ context.Context, objectPath string, data []byte, _ string) error {
	full := filepath.Join(s.root, objectPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("local write mkdir %s: %w", filepath.Dir(full), err)
	}
	return os.WriteFile(full, data, 0o644)
}

func (s *localStore) Read(_ context.Context, objectPath string) ([]byte, error) {
	full := filepath.Join(s.root, objectPath)
	data, err := os.ReadFile(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("local read %s: %w", objectPath, ErrObjectNotFound)
	}
	return data, err
}

func (s *localStore) List(_ context.Context, prefix string) ([]string, error) {
	dir := filepath.Join(s.root, prefix)
	walkRoot := dir
	needFilter := false
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		walkRoot = filepath.Dir(dir)
		needFilter = true
	}
	var out []string
	err := filepath.Walk(walkRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(s.root, p)
		rel = filepath.ToSlash(rel)
		if !needFilter || strings.HasPrefix(rel, prefix) {
			out = append(out, rel)
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *localStore) Close() error { return nil }
