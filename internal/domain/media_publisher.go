package domain

import (
	"media_ads/internal/entities"
	"media_ads/internal/repository"
)

type MediaPublisher struct {
	objectLib ObjectLibraryInterface
	mediaRepo *repository.MediaPublisherRepo
}

type MediaPublisherInterface interface {
	SaveMediaRequest(userID string, mediaID string, title string, description string) (entities.SaveMediaResponse, error)
	UpdateMediaUploadCompleted(mediaID string, objectID string, contentType string) error
	UpdateMediaUploadFailed(mediaID string) error
}

func NewMediaPublisher(objectLibraryAPI ObjectLibraryInterface, mediaPublisherRepo *repository.MediaPublisherRepo) MediaPublisherInterface {
	return &MediaPublisher{
		objectLib: objectLibraryAPI,
		mediaRepo: mediaPublisherRepo,
	}
}

func (m *MediaPublisher) SaveMediaRequest(userID string, mediaID string, title string, description string) (entities.SaveMediaResponse, error) {

	uploadURL, err := m.objectLib.ReserveUploadSlot()
	if err != nil {
		return entities.SaveMediaResponse{}, err
	}

	tx, err := m.mediaRepo.GetDB().Begin()
	if err != nil {
		return entities.SaveMediaResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = m.mediaRepo.SaveMedia(tx, mediaID, title, description, "", "")
	if err != nil {
		return entities.SaveMediaResponse{}, err
	}

	err = m.mediaRepo.SaveMediaOwner(tx, mediaID, userID)
	if err != nil {
		return entities.SaveMediaResponse{}, err
	}

	return entities.SaveMediaResponse{
		MediaID:   mediaID,
		UploadURL: uploadURL,
	}, nil
}

func (m *MediaPublisher) UpdateMediaUploadCompleted(mediaID string, objectID string, contentType string) error {

	return m.mediaRepo.UpdateObjectIDToMedia(nil, mediaID, objectID, contentType)
}

func (m *MediaPublisher) UpdateMediaUploadFailed(mediaID string) error {
	return m.mediaRepo.DeleteMedia(nil, mediaID)
}
