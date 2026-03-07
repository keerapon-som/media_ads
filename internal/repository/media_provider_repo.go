package repository

import (
	"database/sql"
	"media_ads/internal/entities"
)

type MediaProviderRepo struct {
	db *sql.DB
}

func NewMediaProviderRepo(db *sql.DB) *MediaProviderRepo {
	return &MediaProviderRepo{db: db}
}

func (m *MediaProviderRepo) SaveMediaProvider(objectID string, bucket string, key string) (*entities.MediaProviderRepo, error) {
	// Implement logic to save media provider information to the database
	return &entities.MediaProviderRepo{}, nil
}

func (m *MediaProviderRepo) GetMediaProviderByID(id string) (*entities.MediaProviderRepo, error) {
	// Implement logic to retrieve media provider by ID from the database
	return &entities.MediaProviderRepo{}, nil
}
