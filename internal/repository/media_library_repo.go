package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"media_ads/internal/entities"
)

type MediaLibraryRepo struct {
	db *sql.DB
}

func NewMediaLibraryRepo(db *sql.DB) *MediaLibraryRepo {
	return &MediaLibraryRepo{db: db}
}

func (m *MediaLibraryRepo) GetDB() *sql.DB {
	return m.db
}

func (m *MediaLibraryRepo) SaveMediaLibrary(tx *sql.Tx, objectID string, key string, filename string, extension string, sizeBytes int64, contentType string, probeData map[string]any) error {
	probeDataJSON, err := json.Marshal(probeData)
	if err != nil {
		return fmt.Errorf("marshal probe_data: %w", err)
	}

	const query = `
		INSERT INTO media_archives (
			object_id,
			key,
			filename,
			extension,
			size_bytes,
			content_type,
			probe_data
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		ON CONFLICT (object_id) DO UPDATE SET
			key = EXCLUDED.key,
			filename = EXCLUDED.filename,
			extension = EXCLUDED.extension,
			size_bytes = EXCLUDED.size_bytes,
			content_type = EXCLUDED.content_type,
			probe_data = EXCLUDED.probe_data
	`

	if tx != nil {
		if _, err := tx.Exec(query, objectID, key, filename, extension, sizeBytes, contentType, probeDataJSON); err != nil {
			return fmt.Errorf("save media archive with tx: %w", err)
		}
		return nil
	}

	if _, err := m.db.Exec(query, objectID, key, filename, extension, sizeBytes, contentType, probeDataJSON); err != nil {
		return fmt.Errorf("save media archive: %w", err)
	}

	return nil
}

func (m *MediaLibraryRepo) GetMediaLibraryByID(objectID string) (*entities.MediaLibraryRepo, error) {
	const query = `
		SELECT
			object_id,
			key,
			filename,
			extension,
			size_bytes,
			content_type,
			probe_data
		FROM media_archives
		WHERE object_id = $1
		LIMIT 1
	`

	var mediaArchive entities.MediaLibraryRepo
	var probeDataJSON []byte

	err := m.db.QueryRow(query, objectID).Scan(
		&mediaArchive.ObjectID,
		&mediaArchive.Key,
		&mediaArchive.Filename,
		&mediaArchive.Extension,
		&mediaArchive.SizeBytes,
		&mediaArchive.ContentType,
		&probeDataJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("media archive not found: object_id=%s", objectID)
		}
		return nil, fmt.Errorf("get media archive by id: %w", err)
	}

	if len(probeDataJSON) > 0 {
		if err := json.Unmarshal(probeDataJSON, &mediaArchive.ProbeData); err != nil {
			return nil, fmt.Errorf("unmarshal probe_data: %w", err)
		}
	} else {
		mediaArchive.ProbeData = map[string]any{}
	}

	return &mediaArchive, nil
}
