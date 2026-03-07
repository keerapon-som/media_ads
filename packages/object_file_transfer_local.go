package packages

import (
	"crypto/sha256"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type ObjectFileTransferLocal struct {
	RootStoragePath string
}

func NewObjectFileTransferLocal(rootStoragePath string) *ObjectFileTransferLocal {
	return &ObjectFileTransferLocal{
		RootStoragePath: rootStoragePath,
	}
}

func (o *ObjectFileTransferLocal) UploadObject(key string, file *multipart.File) error {
	baseDir := o.RootStoragePath

	storedPath := filepath.Join(baseDir, "/", key)

	// Ensure the base directory exists
	if err := os.MkdirAll(filepath.Dir(storedPath), os.ModePerm); err != nil {
		return err
	}

	dst, err := os.Create(storedPath)
	if err != nil {
		return err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	_, err = io.Copy(writer, *file)
	closeErr := dst.Close()
	if err != nil {
		_ = os.Remove(storedPath)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(storedPath)
		return closeErr
	}

	// checksum := hex.EncodeToString(hasher.Sum(nil))
	// // You can log or store the checksum if needed
	// _ = uploadedBytes
	// _ = checksum

	return nil
}

func (o *ObjectFileTransferLocal) DownloadObject(key string) (*os.File, error) {
	storedPath := filepath.Join(o.RootStoragePath, "/", key)

	return os.Open(storedPath)
}
