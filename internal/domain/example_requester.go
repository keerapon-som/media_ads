package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type UploadResponse struct {
	UploadID        string  `json:"upload_id"`
	Filename        string  `json:"filename"`
	UploadedBytes   int64   `json:"uploaded_bytes"`
	ProgressPercent float64 `json:"progress_percent"`
	Completed       bool    `json:"completed"`
	StoredPath      string  `json:"stored_path"`
	ChecksumSHA256  string  `json:"checksum_sha256"`
	Error           string  `json:"error"`
}

func Use() {
	resp, err := UploadFile("http://localhost:8888", "./sample.mp4")
	if err != nil {
		panic(err)
	}

	fmt.Printf("upload_id=%s completed=%v stored_path=%s checksum=%s\n",
		resp.UploadID, resp.Completed, resp.StoredPath, resp.ChecksumSHA256)
}

func UploadFile(apiBaseURL, filePath string) (*UploadResponse, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	filename := filepath.Base(filePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("filename", filename); err != nil {
		return nil, fmt.Errorf("write filename: %w", err)
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiBaseURL+"/upload_media", body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var out UploadResponse
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w body=%s", err, string(respBytes))
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upload failed status=%d error=%s", resp.StatusCode, out.Error)
	}

	return &out, nil
}

func GetUploadStatus(apiBaseURL, uploadID string) (*UploadResponse, error) {
	url := fmt.Sprintf("%s/upload_media/%s/status", apiBaseURL, uploadID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status check failed: %d (%s)", resp.StatusCode, out.Error)
	}
	return &out, nil
}
