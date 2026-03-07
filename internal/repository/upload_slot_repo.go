package repository

import (
	"database/sql"
	"media_ads/internal/entities"
)

func (m *ObjectLibraryRepo) ReserveUploadSlot(tx *sql.Tx, uploadID string) error {
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

func (m *ObjectLibraryRepo) UpdateUploadStatus(tx *sql.Tx, uploadID string, status string) error {

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

func (m *ObjectLibraryRepo) GetUploadSlot(uploadID string) (*entities.UploadSlotRepo, error) {
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
