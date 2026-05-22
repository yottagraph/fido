package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const checkpointObjectPath = "checkpoints/checkpoint.json"

// Checkpoint persists how far the last successful Run got. Extend the
// shape (add `LastBlock int64`, `LastCursor string`, …) to match what
// the upstream source requires to resume.
type Checkpoint struct {
	Cursor string `json:"cursor,omitempty"`
}

// LoadCheckpoint reads the persisted checkpoint, returning a zero
// Checkpoint when none exists.
func LoadCheckpoint(ctx context.Context, store Store) (*Checkpoint, error) {
	data, err := store.Read(ctx, checkpointObjectPath)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return &Checkpoint{}, nil
		}
		return nil, err
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint: parse: %w", err)
	}
	return &cp, nil
}

// SaveCheckpoint persists the checkpoint to the output store.
func SaveCheckpoint(ctx context.Context, store Store, cp *Checkpoint) error {
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return store.Write(ctx, checkpointObjectPath, data, "application/json")
}
