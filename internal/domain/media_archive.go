package domain

import (
	"context"
	"encoding/json"
	"media_ads/internal/entities"
	"media_ads/internal/repository"
	"media_ads/packages"
	"mime/multipart"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type MediaArchive struct {
	ObjectFileTransfer *packages.ObjectFileTransferLocal
	mediaProviderRepo  *repository.MediaArchiveRepo
}

func NewMediaArchive(objectFileTransfer *packages.ObjectFileTransferLocal, mediaProviderRepo *repository.MediaArchiveRepo) *MediaArchive {
	return &MediaArchive{
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

func (m *MediaArchive) UploadMedia(objectID string, fileHeader *multipart.FileHeader) error {

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	inspection := collectMediaInspection(fileHeader)

	key := "subfolder" + "/" + objectID + "." + inspection.Extension

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

	err = m.mediaProviderRepo.SaveMediaArchive(
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

	return nil
}

func (m *MediaArchive) DownloadMedia(objectID string) (*entities.DownloadResponse, error) {

	mediaProvider, err := m.mediaProviderRepo.GetMediaArchiveByID(objectID)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	file, err := m.ObjectFileTransfer.DownloadObject(mediaProvider.Key)
	if err != nil {
		return &entities.DownloadResponse{}, err
	}

	return &entities.DownloadResponse{
		Filename:  mediaProvider.Filename,
		Extension: mediaProvider.Extension,
		SizeBytes: mediaProvider.SizeBytes,
		File:      file,
	}, nil
}
