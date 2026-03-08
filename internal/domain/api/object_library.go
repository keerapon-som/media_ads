package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"media_ads/internal/domain"
	"media_ads/internal/entities"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"strings"
	"time"
)

type ObjectLibraryAPI struct {
	baseurl    string
	httpClient *http.Client
	secureKey  string
}

func NewObjectLibraryAPI(baseurl string, httpClient *http.Client, secureKey string) domain.ObjectLibraryInterface {
	baseURL := strings.TrimRight(baseurl, "/")
	if baseURL == "" {
		panic("baseurl is required for ObjectLibraryAPI")
	}

	return &ObjectLibraryAPI{
		baseurl: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		secureKey: secureKey,
	}
}

func (o *ObjectLibraryAPI) ReserveUploadSlot() (string, error) {
	resp, err := o.doRequest(http.MethodPost, "/object_library/researve_upload_slot", true, nil, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := ensureSuccess(resp); err != nil {
		return "", err
	}

	var payload struct {
		UploadID string `json:"upload_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode reserve upload slot response: %w", err)
	}

	if payload.UploadID == "" {
		return "", fmt.Errorf("reserve upload slot response missing upload_id")
	}

	return payload.UploadID, nil
}

func (o *ObjectLibraryAPI) UploadObject(uploadID string, objectID string, fileHeader *multipart.FileHeader) error {
	if uploadID == "" {
		return fmt.Errorf("upload_id is required")
	}
	if fileHeader == nil {
		return fmt.Errorf("file header is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open multipart file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if objectID != "" {
		if err := writer.WriteField("object_id", objectID); err != nil {
			return fmt.Errorf("write object_id field: %w", err)
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileHeader.Filename))
	h.Set("Content-Type", fileHeader.Header.Get("Content-Type"))

	part, err := writer.CreatePart(h)
	if err != nil {
		return fmt.Errorf("create file part: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("finalize multipart body: %w", err)
	}

	resp, err := o.doRequest(http.MethodPut, "/object_library/upload/"+uploadID, false, body, writer.FormDataContentType())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := ensureSuccess(resp); err != nil {
		return err
	}

	return nil
}

func (o *ObjectLibraryAPI) GetObject(objectID string) (*entities.DownloadResponse, error) {
	if objectID == "" {
		return nil, fmt.Errorf("object_id is required")
	}

	resp, err := o.doRequest(http.MethodGet, "/object_library/object/"+objectID, false, nil, "")
	if err != nil {
		return nil, err
	}

	if err := ensureSuccess(resp); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	extension := fileExtensionFromHeaders(resp.Header.Get("Content-Disposition"), contentType)

	tmpFile, err := os.CreateTemp("", "object-library-*")
	if err != nil {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("create temp file for download: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = resp.Body.Close()
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("read object response body: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("close object response body: %w", err)
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("rewind downloaded file: %w", err)
	}

	return &entities.DownloadResponse{
		Filename:    objectID,
		Extension:   extension,
		SizeBytes:   responseContentLength(resp),
		File:        tmpFile,
		ContentType: contentType,
	}, nil
}

func (o *ObjectLibraryAPI) GetObjectInfo(objectID string) (*entities.ObjectLibraryRepo, error) {
	if objectID == "" {
		return nil, fmt.Errorf("object_id is required")
	}

	resp, err := o.doRequest(http.MethodGet, "/object_library/object_info/"+objectID, false, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ensureSuccess(resp); err != nil {
		return nil, err
	}

	var payload struct {
		Status string                     `json:"status"`
		Data   entities.ObjectLibraryRepo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode get object info response: %w", err)
	}

	return &payload.Data, nil
}

func (o *ObjectLibraryAPI) DeleteObject(objectID string) error {
	if objectID == "" {
		return fmt.Errorf("object_id is required")
	}

	resp, err := o.doRequest(http.MethodDelete, "/object_library/object/"+objectID, false, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ensureSuccess(resp)
}

func (o *ObjectLibraryAPI) PublishObject(objectID string) error {
	if objectID == "" {
		return fmt.Errorf("object_id is required")
	}

	resp, err := o.doRequest(http.MethodPost, "/object_library/publish/"+objectID, true, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ensureSuccess(resp)
}

func (o *ObjectLibraryAPI) UnpublishObject(objectID string) error {
	if objectID == "" {
		return fmt.Errorf("object_id is required")
	}

	resp, err := o.doRequest(http.MethodPost, "/object_library/unpublish/"+objectID, true, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ensureSuccess(resp)
}

func (o *ObjectLibraryAPI) doRequest(method, endpoint string, internalOnly bool, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, o.baseurl+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, endpoint, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if internalOnly {
		if strings.TrimSpace(o.secureKey) == "" {
			return nil, fmt.Errorf("OBJECT_LIBRARY_INTERNAL_TOKEN is required for internal endpoint %s", endpoint)
		}
		req.Header.Set("X-Internal-Token", o.secureKey)
		req.Header.Set("Authorization", "Bearer "+o.secureKey)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request %s %s: %w", method, endpoint, err)
	}

	return resp, nil
}

func ensureSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	bodyText := strings.TrimSpace(string(bodyBytes))
	if bodyText == "" {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, bodyText)
}

func responseContentLength(resp *http.Response) int64 {
	if resp.ContentLength > 0 {
		return resp.ContentLength
	}
	return 0
}

func fileExtensionFromHeaders(contentDisposition string, contentType string) string {
	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if filename := params["filename"]; filename != "" {
			ext := strings.TrimPrefix(path.Ext(filename), ".")
			if ext != "" {
				return ext
			}
		}
	}

	if contentType != "" {
		if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
			return strings.TrimPrefix(exts[0], ".")
		}
	}

	return "bin"
}
