package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"media_ads/internal/entities"
	"media_ads/internal/repository"
	"media_ads/packages"
	"mime/multipart"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ObjectLibrary struct {
	ObjectFileTransfer *packages.ObjectFileTransferLocal
	mediaProviderRepo  *repository.ObjectLibraryRepo
}

func NewObjectLibrary(objectFileTransfer *packages.ObjectFileTransferLocal, mediaProviderRepo *repository.ObjectLibraryRepo) *ObjectLibrary {
	return &ObjectLibrary{
		ObjectFileTransfer: objectFileTransfer,
		mediaProviderRepo:  mediaProviderRepo,
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

func (m *ObjectLibrary) UploadObject(upload_id string, objectID string, fileHeader *multipart.FileHeader) error {

	tx, err := m.mediaProviderRepo.GetDB().Begin()
	if err != nil {
		return err
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
		return err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	inspection := collectMediaInspection(fileHeader)

	key := "subfolder" + "/" + objectID

	err = m.mediaProviderRepo.ObjectLibRepo.SaveObjectLibrary(
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
		return err
	}

	err = m.ObjectFileTransfer.UploadObject(key, &file)
	if err != nil {
		return err
	}

	err = m.updateUploadSlotStatus(tx, upload_id, "completed")
	if err != nil {
		return err
	}

	return nil
}

func (m *ObjectLibrary) GetObject(objectID string) (*entities.DownloadResponse, error) {

	mediaProvider, err := m.mediaProviderRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	file, err := m.ObjectFileTransfer.GetObject(mediaProvider.Key)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	return &entities.DownloadResponse{
		Filename:    mediaProvider.Filename,
		Extension:   mediaProvider.Extension,
		SizeBytes:   mediaProvider.SizeBytes,
		File:        file,
		ContentType: mediaProvider.ContentType,
	}, nil
}

func (m *ObjectLibrary) GetObjectInfo(objectID string) (*entities.ObjectLibraryRepo, error) {

	return m.mediaProviderRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
}

func (m *ObjectLibrary) DeleteObject(objectID string) error {
	mediaProvider, err := m.mediaProviderRepo.ObjectLibRepo.GetObjectLibraryByID(objectID)
	if err != nil {
		return err
	}

	tx, err := m.mediaProviderRepo.GetDB().Begin()
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

	err = m.mediaProviderRepo.ObjectLibRepo.DeleteObjectLibraryByObjectID(tx, objectID)
	if err != nil {
		return err
	}

	err = m.ObjectFileTransfer.DeleteObject(mediaProvider.Key)
	if err != nil {
		return err
	}

	return nil
}

func (m *ObjectLibrary) ReserveUploadSlot() (string, error) {

	uploadID := uuid.NewString()

	return uploadID, m.mediaProviderRepo.UploadSlotRepo.ReserveUploadSlot(nil, uploadID)
}

// func (m *ObjectLibrary) updateUploadCompletion(tx *sql.Tx, uploadID string, success bool) error {
// 	return m.mediaProviderRepo.UpdateUploadCompletion(tx, uploadID, success)
// }

func (m *ObjectLibrary) claimUploadSlot(tx *sql.Tx, uploadID string) error {

	resp, err := m.mediaProviderRepo.UploadSlotRepo.GetUploadSlot(uploadID)
	if err != nil {
		return err
	}

	if resp.Status != "pending" {
		return fmt.Errorf("upload slot is not pending: upload_id=%s, status=%s", uploadID, resp.Status)
	}

	return m.mediaProviderRepo.UploadSlotRepo.UpdateUploadStatus(tx, uploadID, "claimed")
}

func (m *ObjectLibrary) updateUploadSlotStatus(tx *sql.Tx, uploadID string, status string) error {

	return m.mediaProviderRepo.UploadSlotRepo.UpdateUploadStatus(tx, uploadID, status)
}
