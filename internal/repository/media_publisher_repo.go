package repository

import (
	"database/sql"
	"time"
)

type MediaPublisherRepo struct {
	db *sql.DB
}

func NewMediaPublisherRepo(db *sql.DB) *MediaPublisherRepo {
	return &MediaPublisherRepo{
		db: db,
	}
}

func (m *MediaPublisherRepo) GetDB() *sql.DB {
	return m.db
}

func (m *MediaPublisherRepo) SaveMedia(tx *sql.Tx, mediaID string, title string, description string, objectID string, contentType string) error {
	const query = `
		INSERT INTO media (
			media_id,
			title,
			description,
			object_id,
			content_type
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	if tx != nil {
		if _, err := tx.Exec(query, mediaID, title, description, objectID, contentType); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, mediaID, title, description, objectID, contentType); err != nil {
		return err
	}
	return nil
}

func (m *MediaPublisherRepo) DeleteMedia(tx *sql.Tx, mediaID string) error {
	const query = `
		DELETE FROM media
		WHERE media_id = $1
	`

	if tx != nil {
		if _, err := tx.Exec(query, mediaID); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, mediaID); err != nil {
		return err
	}
	return nil
}

func (m *MediaPublisherRepo) UpdateObjectIDToMedia(tx *sql.Tx, mediaID string, objectID string, contentType string) error {
	const query = `
		UPDATE media
		SET object_id = $1,
			content_type = $2,
			updated_at = NOW()
		WHERE media_id = $3
	`

	if tx != nil {
		if _, err := tx.Exec(query, objectID, contentType, mediaID); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, objectID, contentType, mediaID); err != nil {
		return err
	}
	return nil
}

func (m *MediaPublisherRepo) SaveMediaOwner(tx *sql.Tx, mediaID string, userID string) error {
	const query = `
		INSERT INTO media_owner (
			media_id,
			user_id
		)
		VALUES ($1, $2)
		ON CONFLICT (media_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			updated_at = NOW()
	`

	if tx != nil {
		if _, err := tx.Exec(query, mediaID, userID); err != nil {
			return err
		}
		return nil
	}

	if _, err := m.db.Exec(query, mediaID, userID); err != nil {
		return err
	}

	return nil
}

func (m *MediaPublisherRepo) SavePublishRegister(tx *sql.Tx, mediaID string, publishFrom time.Time, publishTo *time.Time) (int64, error) {
	const query = `
		INSERT INTO media_publish_register (
			media_id,
			publish_from,
			publish_to
		)
		VALUES ($1, $2, $3)
		RETURNING publisher_id
	`

	var publisherID int64

	if tx != nil {
		if err := tx.QueryRow(query, mediaID, publishFrom, publishTo).Scan(&publisherID); err != nil {
			return 0, err
		}
		return publisherID, nil
	}

	if err := m.db.QueryRow(query, mediaID, publishFrom, publishTo).Scan(&publisherID); err != nil {
		return 0, err
	}

	return publisherID, nil
}

func (m *MediaPublisherRepo) GetCurrentPublishWindow(mediaID string, now time.Time) (*time.Time, *time.Time, error) {
	const query = `
		SELECT
			publish_from,
			publish_to
		FROM media_publish_register
		WHERE media_id = $1
			AND publish_from <= $2
			AND (publish_to IS NULL OR $2 < publish_to)
		ORDER BY publish_from DESC
		LIMIT 1
	`

	var publishFrom time.Time
	var publishTo sql.NullTime

	err := m.db.QueryRow(query, mediaID, now).Scan(&publishFrom, &publishTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var publishToPtr *time.Time
	if publishTo.Valid {
		publishToPtr = &publishTo.Time
	}

	return &publishFrom, publishToPtr, nil
}
