package uploader

import "io"

type CloudUploader interface {
	Upload(io.Reader, string) error
}

func UploadToCloud(cloudUploader CloudUploader, file io.Reader, path string) error {
	if cloudUploader != nil {
		return cloudUploader.Upload(file, path)
	}
	return nil
}
