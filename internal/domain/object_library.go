package domain

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"media_ads/internal/entities"
	"media_ads/internal/repository"
	"media_ads/packages"
	"mime/multipart"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ObjectLibrary struct {
	ObjectFileTransfer              *packages.ObjectFileTransferLocal
	objectLibraryRepo               *repository.ObjectLibraryRepo
	defaultCallbackURLUploadSuccess string
}

type ObjectLibraryInterface interface {
	ReserveUploadSlot() (string, error)
	UploadObject(upload_id string, objectID string, fileHeader *multipart.FileHeader) (entities.MediaInfo, error)
	GetObject(objectID string) (*entities.DownloadResponse, error)
	GetObjectInfo(objectID string) (*entities.ObjectLibraryRepo, error)
	DeleteObject(objectID string) error
	PublishObject(objectID string) error
	UnpublishObject(objectID string) error
	CallbackUpdateUploadSuccess(mediaID string, objectID string, contentType string, isSuccess bool, callbackURL string) error
}

func NewObjectLibrary(defaultCallbackURLUploadSuccess string, objectFileTransfer *packages.ObjectFileTransferLocal, objectLibraryRepo *repository.ObjectLibraryRepo) ObjectLibraryInterface {
	return &ObjectLibrary{
		defaultCallbackURLUploadSuccess: defaultCallbackURLUploadSuccess,
		ObjectFileTransfer:              objectFileTransfer,
		objectLibraryRepo:               objectLibraryRepo,
	}
}

func collectMediaInspection(fileHeader *multipart.FileHeader) entities.MediaInfo {

	probeData, probeErr := ffprobeFromFileHeader(fileHeader)
	if probeErr != nil {
		return entities.MediaInfo{
			Filename:    fileHeader.Filename,
			Extension:   strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), "."),
			SizeBytes:   fileHeader.Size,
			ContentType: fileHeader.Header.Get("Content-Type"),
			ProbeData: map[string]any{
				"ffprobe_error": probeErr.Error(),
			},
		}
	}

	return entities.MediaInfo{
		Filename:    fileHeader.Filename,
		Extension:   strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), "."),
		SizeBytes:   fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		ProbeData:   probeData,
	}

}

func ffprobeFromFileHeader(fileHeader *multipart.FileHeader) (map[string]any, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "error",
		"-show_streams",
		"-show_format",
		"-print_format", "json",
		"-i", "pipe:0",
	)
	cmd.Stdin = src

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var parsed map[string]any
	if err := json.Unmarshal(output, &parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func (m *ObjectLibrary) UploadObject(upload_id string, objectID string, fileHeader *multipart.FileHeader) (entities.MediaInfo, error) {

	tx, err := m.objectLibraryRepo.GetDB().Begin()
	if err != nil {
		return entities.MediaInfo{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			m.updateUploadSlotStatus(nil, upload_id, "failed")
		} else {
			err = tx.Commit()

		}
	}()

	err = m.claimUploadSlot(tx, upload_id)
	if err != nil {
		return entities.MediaInfo{}, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return entities.MediaInfo{}, err
	}
	defer file.Close()

	inspection := collectMediaInspection(fileHeader)

	key := "subfolder" + "/" + objectID

	err = m.objectLibraryRepo.ObjectLibRepo.SaveObjectLibrary(
		tx,
		objectID,
		key,
		inspection.Filename,
		inspection.Extension,
		inspection.SizeBytes,
		inspection.ContentType,
		inspection.ProbeData,
	)
	if err != nil {
		return entities.MediaInfo{}, err
	}

	err = m.ObjectFileTransfer.UploadObject(key, &file)
	if err != nil {
		return entities.MediaInfo{}, err
	}

	err = m.updateUploadSlotStatus(tx, upload_id, "completed")
	if err != nil {
		return entities.MediaInfo{}, err
	}

	return inspection, nil
}

func (m *ObjectLibrary) GetObject(objectID string) (*entities.DownloadResponse, error) {

	objectLibrary, err := m.objectLibraryRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	if !objectLibrary.IsPublished {
		return &entities.DownloadResponse{}, fmt.Errorf("object is not published: object_id=%s", objectID)
	}

	file, err := m.ObjectFileTransfer.GetObject(objectLibrary.Key)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	return &entities.DownloadResponse{
		Filename:    objectLibrary.Filename,
		Extension:   objectLibrary.Extension,
		SizeBytes:   objectLibrary.SizeBytes,
		File:        file,
		ContentType: objectLibrary.ContentType,
	}, nil
}

func (m *ObjectLibrary) GetObjectInfo(objectID string) (*entities.ObjectLibraryRepo, error) {

	return m.objectLibraryRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
}

func (m *ObjectLibrary) DeleteObject(objectID string) error {
	objectLibrary, err := m.objectLibraryRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
	if err != nil {
		return err
	}

	tx, err := m.objectLibraryRepo.GetDB().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = m.objectLibraryRepo.ObjectLibRepo.DeleteObjectLibraryByObjectID(tx, objectID)
	if err != nil {
		return err
	}

	err = m.ObjectFileTransfer.DeleteObject(objectLibrary.Key)
	if err != nil {
		return err
	}

	return nil
}

func (m *ObjectLibrary) ReserveUploadSlot() (string, error) {

	uploadID := uuid.NewString()

	return uploadID, m.objectLibraryRepo.UploadSlotRepo.ReserveUploadSlot(nil, uploadID)
}

func (m *ObjectLibrary) PublishObject(objectID string) error {

	return m.objectLibraryRepo.ObjectLibRepo.UpdatePublishedStatus(nil, objectID, true)
}

func (m *ObjectLibrary) UnpublishObject(objectID string) error {

	return m.objectLibraryRepo.ObjectLibRepo.UpdatePublishedStatus(nil, objectID, false)
}

// func (m *ObjectLibrary) updateUploadCompletion(tx *sql.Tx, uploadID string, success bool) error {
// 	return m.objectLibraryRepo.UpdateUploadCompletion(tx, uploadID, success)
// }

func (m *ObjectLibrary) claimUploadSlot(tx *sql.Tx, uploadID string) error {

	resp, err := m.objectLibraryRepo.UploadSlotRepo.GetUploadSlot(uploadID)
	if err != nil {
		return err
	}

	if resp.Status != "pending" {
		return fmt.Errorf("upload slot is not pending: upload_id=%s, status=%s", uploadID, resp.Status)
	}

	return m.objectLibraryRepo.UploadSlotRepo.UpdateUploadStatus(tx, uploadID, "claimed")
}

func (m *ObjectLibrary) updateUploadSlotStatus(tx *sql.Tx, uploadID string, status string) error {

	return m.objectLibraryRepo.UploadSlotRepo.UpdateUploadStatus(tx, uploadID, status)
}

// h.mediaPublisher.UpdateMediaUploadCompleted(
// 	mediaID,
// 	objectID,
// 	mediaInfo.ContentType,
// )

func (m *ObjectLibrary) CallbackUpdateUploadSuccess(mediaID string, objectID string, contentType string, isSuccess bool, callbackURL string) error {

	if callbackURL == "" {
		callbackURL = m.defaultCallbackURLUploadSuccess
	}

	payload := &entities.CallbackUpdateUploadStatusRequest{
		MediaID:     mediaID,
		ObjectID:    objectID,
		ContentType: contentType,
		Success:     isSuccess,
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, callbackURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
