package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"media_ads/internal/entities"
)

type ObjectLibRepo struct {
	db *sql.DB
}

type ObjectLibraryRepo struct {
	db             *sql.DB
	UploadSlotRepo *UploadSlotRepo
	ObjectLibRepo  *ObjectLibRepo
}

func NewObjectLibraryRepo(db *sql.DB) *ObjectLibraryRepo {

	uploadSlotRepo := &UploadSlotRepo{db: db}
	objectLibRepo := &ObjectLibRepo{db: db}

	return &ObjectLibraryRepo{
		db:             db,
		UploadSlotRepo: uploadSlotRepo,
		ObjectLibRepo:  objectLibRepo,
	}
}

func (m *ObjectLibraryRepo) GetDB() *sql.DB {
	return m.db
}

func (m *ObjectLibRepo) SaveObjectLibrary(tx *sql.Tx, objectID string, key string, filename string, extension string, sizeBytes int64, contentType string, probeData map[string]any) error {
	probeDataJSON, err := json.Marshal(probeData)
	if err != nil {
		return fmt.Errorf("marshal probe_data: %w", err)
	}

	const query = `
		INSERT INTO object_library (
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
			probe_data = EXCLUDED.probe_data,
			updated_at = NOW()
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

func (m *ObjectLibRepo) GetObjectLibraryByID(objectID string) (*entities.ObjectLibraryRepo, error) {
	const query = `
		SELECT
			object_id,
			key,
			filename,
			extension,
			size_bytes,
			content_type,
			probe_data,
			is_published,
			created_at,
			updated_at
		FROM object_library
		WHERE object_id = $1
		LIMIT 1
	`

	var mediaArchive entities.ObjectLibraryRepo
	var probeDataJSON []byte

	err := m.db.QueryRow(query, objectID).Scan(
		&mediaArchive.ObjectID,
		&mediaArchive.Key,
		&mediaArchive.Filename,
		&mediaArchive.Extension,
		&mediaArchive.SizeBytes,
		&mediaArchive.ContentType,
		&probeDataJSON,
		&mediaArchive.IsPublished,
		&mediaArchive.CreatedAt,
		&mediaArchive.UpdatedAt,
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

func (m *ObjectLibRepo) DeleteObjectLibraryByObjectID(tx *sql.Tx, objectID string) error {
	const query = `
		DELETE FROM object_library
		WHERE object_id = $1
	`

	if tx != nil {
		if _, err := tx.Exec(query, objectID); err != nil {
			return fmt.Errorf("delete media archive with tx: %w", err)
		}
		return nil
	}

	if _, err := m.db.Exec(query, objectID); err != nil {
		return fmt.Errorf("delete media archive: %w", err)
	}

	return nil
}

func (m *ObjectLibRepo) UpdatePublishedStatus(tx *sql.Tx, objectID string, isPublished bool) error {
	const query = `
		UPDATE object_library
		SET is_published = $1, updated_at = NOW()
		WHERE object_id = $2
	`

	if tx != nil {
		result, err := tx.Exec(query, isPublished, objectID)
		if err != nil {
			return fmt.Errorf("update published status with tx: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("update published status with tx rows affected: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("object library not found for update: object_id=%s", objectID)
		}
		return nil
	}

	result, err := m.db.Exec(query, isPublished, objectID)
	if err != nil {
		return fmt.Errorf("update published status: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update published status rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("object library not found for update: object_id=%s", objectID)
	}

	return nil
}

type UploadSlotRepo struct {
	db *sql.DB
}

func (m *UploadSlotRepo) ReserveUploadSlot(tx *sql.Tx, uploadID string) error {
	const query = `
		INSERT INTO upload_slot (
			upload_id,
			status
		)
		VALUES ($1, 'pending')
	`

	if tx != nil {
		if _, err := tx.Exec(query, uploadID); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, uploadID); err != nil {
		return err
	}

	return nil
}

func (m *UploadSlotRepo) UpdateUploadStatus(tx *sql.Tx, uploadID string, status string) error {

	const query = `
		UPDATE upload_slot
		SET status = $1, updated_at = NOW()
		WHERE upload_id = $2
	`

	if tx != nil {
		if _, err := tx.Exec(query, status, uploadID); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, status, uploadID); err != nil {
		return err
	}

	return nil
}

func (m *UploadSlotRepo) GetUploadSlot(uploadID string) (*entities.UploadSlotRepo, error) {
	const query = `
		SELECT upload_id, status, created_at, updated_at
		FROM upload_slot
		WHERE upload_id = $1
	`

	row := m.db.QueryRow(query, uploadID)

	var slot entities.UploadSlotRepo
	err := row.Scan(&slot.UploadID, &slot.Status, &slot.CreatedAt, &slot.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &slot, nil
}
