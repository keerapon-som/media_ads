package domain

import (
	"context"
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
)

type MediaProvider struct {
	bucket             string
	ObjectFileTransfer *packages.ObjectFileTransferLocal
	mediaProviderRepo  *repository.MediaProviderRepo
}

func NewMediaProvider(bucket string, objectFileTransfer *packages.ObjectFileTransferLocal, mediaProviderRepo *repository.MediaProviderRepo) *MediaProvider {
	return &MediaProvider{
		bucket:             bucket,
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

func (m *MediaProvider) UploadMedia(objectID, userID string, file *multipart.File, fileHeader *multipart.FileHeader) error {

	inspection := collectMediaInspection(fileHeader)

	fmt.Println("FileName ", inspection.Filename)

	key := userID + "/" + objectID + "." + inspection.Extension

	err := m.ObjectFileTransfer.UploadObject(m.bucket, key, file)
	if err != nil {
		return err
	}

	_, err = m.mediaProviderRepo.SaveMediaProvider(objectID, m.bucket, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m *MediaProvider) GetMedia(id string) (*[]byte, error) {
	mediaProvider, err := m.mediaProviderRepo.GetMediaProviderByID(id)
	if err != nil {
		return nil, err
	}

	b, err := m.ObjectFileTransfer.GetObject(mediaProvider.Bucket, mediaProvider.Key)
	if err != nil {
		return nil, err
	}

	return b, nil
}
