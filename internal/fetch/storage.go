package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
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
//   - gs://bucket           → GCS bucket root (production)
//   - gs://bucket/some/path → GCS bucket with an in-bucket prefix
//   - file:///abs/path      → local (dev / tests)
//   - /abs/path             → local (dev / tests)
//
// The GCS backend uses Application Default Credentials at runtime — on
// Cloud Run that's the job's runtime service account, locally it's
// whatever `gcloud auth application-default login` set.
func NewStoreFromURI(ctx context.Context, uri string) (Store, error) {
	switch {
	case strings.HasPrefix(uri, "gs://"):
		return newGCSStore(ctx, uri)
	case strings.HasPrefix(uri, "file://"):
		return newLocalStore(strings.TrimPrefix(uri, "file://"))
	case strings.HasPrefix(uri, "/"):
		return newLocalStore(uri)
	default:
		return nil, fmt.Errorf("storage: unsupported URI %q (want gs://, file://, or absolute path)", uri)
	}
}

// gcsStore is the Store backed by Google Cloud Storage. One client per
// store, shared across all calls; closed by Close.
type gcsStore struct {
	client *storage.Client
	bucket string
	prefix string // in-bucket prefix; empty if the URI was bare gs://bucket
}

func newGCSStore(ctx context.Context, uri string) (*gcsStore, error) {
	rest := strings.TrimSuffix(strings.TrimPrefix(uri, "gs://"), "/")
	bucket := rest
	prefix := ""
	if i := strings.Index(rest, "/"); i >= 0 {
		bucket = rest[:i]
		prefix = rest[i+1:]
	}
	if bucket == "" {
		return nil, fmt.Errorf("storage: %q has no bucket name", uri)
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: new gcs client: %w", err)
	}
	return &gcsStore{client: client, bucket: bucket, prefix: prefix}, nil
}

func (s *gcsStore) Root() string {
	if s.prefix == "" {
		return "gs://" + s.bucket
	}
	return "gs://" + s.bucket + "/" + s.prefix
}

// objectName resolves a store-relative path to a full object name in
// the underlying bucket, accounting for the optional in-bucket prefix.
func (s *gcsStore) objectName(objectPath string) string {
	p := strings.TrimLeft(objectPath, "/")
	if s.prefix == "" {
		return p
	}
	return s.prefix + "/" + p
}

func (s *gcsStore) Write(ctx context.Context, objectPath string, data []byte, contentType string) error {
	name := s.objectName(objectPath)
	w := s.client.Bucket(s.bucket).Object(name).NewWriter(ctx)
	if contentType != "" {
		w.ContentType = contentType
	}
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return fmt.Errorf("gcs write gs://%s/%s: %w", s.bucket, name, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("gcs close gs://%s/%s: %w", s.bucket, name, err)
	}
	return nil
}

func (s *gcsStore) Read(ctx context.Context, objectPath string) ([]byte, error) {
	name := s.objectName(objectPath)
	r, err := s.client.Bucket(s.bucket).Object(name).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, fmt.Errorf("gcs read gs://%s/%s: %w", s.bucket, name, ErrObjectNotFound)
		}
		return nil, fmt.Errorf("gcs read gs://%s/%s: %w", s.bucket, name, err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("gcs read gs://%s/%s: %w", s.bucket, name, err)
	}
	return data, nil
}

func (s *gcsStore) List(ctx context.Context, prefix string) ([]string, error) {
	fullPrefix := s.objectName(prefix)
	it := s.client.Bucket(s.bucket).Objects(ctx, &storage.Query{Prefix: fullPrefix})
	var out []string
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gcs list gs://%s/%s: %w", s.bucket, fullPrefix, err)
		}
		// Strip the store's in-bucket prefix so callers see paths
		// relative to the store root, matching localStore semantics.
		name := attrs.Name
		if s.prefix != "" {
			name = strings.TrimPrefix(name, s.prefix+"/")
		}
		out = append(out, name)
	}
	return out, nil
}

func (s *gcsStore) Close() error {
	return s.client.Close()
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
