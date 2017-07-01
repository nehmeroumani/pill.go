package uploader

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"image/color"

	"github.com/nehmeroumani/izero"

	"github.com/nehmeroumani/pill.go/clean"
	"github.com/nehmeroumani/pill.go/uploader/gcs"
)

var (
	imageExtensions   = []string{".jpeg", ".jpg", ".gif", ".png"}
	imageContentTypes = []string{"image/jpeg", "image/jpg", "image/gif", "image/png"}
	imageSizes        map[string]map[string][]uint

	pdfContentTypes = []string{"application/pdf", "application/x-pdf", "application/acrobat", "applications/vnd.pdf", "text/pdf", "text/x-pdf"}

	documentExtensions   = []string{".doc", ".dot", ".docx", ".dotx", ".docm", ".dotm"}
	documentContentTypes = []string{"application/zip", "application/msword", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.openxmlformats-officedocument.wordprocessingml.template", "application/vnd.ms-word.document.macroEnabled.12", "application/vnd.ms-word.template.macroEnabled.12"}

	svgExtensions   = []string{".svg", ".svgz"}
	svgContentTypes = []string{"image/svg+xml", "text/xml", "text/xml; charset=utf-8"}

	baseLocalPath, baseCloudPath string
)

func InitUploader(BaseLocalPath string, BaseCloudPath string, imgSizes map[string]map[string][]uint) {
	imageSizes = imgSizes
	baseLocalPath = filepath.FromSlash(BaseLocalPath)
	baseCloudPath = filepath.FromSlash(BaseCloudPath)
}

type MultipleUpload struct {
	FormData        *multipart.Form
	FilesInputName  string
	FileType        string
	Urls            []string
	ImageSizes      []string
	ImageCategory   string
	localPath       string
	cloudPath       string
	WithCrop        bool
	BackgroundColor *color.RGBA
}

func (this *MultipleUpload) Upload() (error, []string) {
	if this.FormData != nil {
		uploadedFilesNames := []string{}
		files := this.FormData.File[this.FilesInputName] // grab the filenames
		for i, _ := range files {
			file, err := files[i].Open()
			defer file.Close()

			if err != nil {
				clean.Error(err)
				return err, nil
			}
			fileExtension := filepath.Ext(files[i].Filename)
			fileExtension = strings.ToLower(fileExtension)

			fileData := make([]byte, 512)
			_, err = file.Read(fileData)
			if err != nil {
				clean.Error(err)
				return err, nil
			}

			isValidFileType, fileType, fileTypeName := isValidFileType(this.FileType, fileData, fileExtension)

			if !isValidFileType {
				return errors.New("invalid_file_type"), nil
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				clean.Error(err)
				return err, nil
			}
			randomFileName := generateRandomFileName(fileExtension)
			if this.cloudPath != "" {
				if err = UploadToCloud(gcs.GetClient(), file, this.GetCloudFilePath(randomFileName)); err == nil {
					if fileTypeName == "image" && this.ImageSizes != nil {
						var resizedImages map[string]*izero.Img
						if this.WithCrop {
							resizedImages, err = izero.ResizeImgWithCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes())
						} else {
							resizedImages, err = izero.ResizeImgWithoutCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.BackgroundColor)
						}
						if err != nil {
							clean.Error(err)
						} else {
							for sizeName, resizedImage := range resizedImages {
								if err = UploadToCloud(gcs.GetClient(), resizedImage.ToReader(), this.GetCloudFilePath(randomFileName, sizeName)); err != nil {
									clean.Error(err)
								}
							}
						}
					}
					uploadedFilesNames = append(uploadedFilesNames, randomFileName)
				} else {
					clean.Error(err)
				}
			} else {
				if ok, pathErr := CreateFolderPath(this.localPath); ok {
					out, err := os.Create(filepath.Join(this.localPath, randomFileName))
					defer out.Close()
					if err != nil {
						clean.Error(errors.New("Unable to create the file for writing. Check your write access privilege : " + err.Error()))
						return err, nil
					}

					_, err = io.Copy(out, file)

					if err != nil {
						clean.Error(err)
						return err, nil
					}
					_, err = file.Seek(0, 0)
					if err != nil {
						clean.Error(err)
						return err, nil
					}
					if fileTypeName == "image" && this.ImageSizes != nil {
						if this.WithCrop {
							izero.ResizeImgWithCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.LocalPath())
						} else {
							izero.ResizeImgWithoutCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.LocalPath(), this.BackgroundColor)
						}
					}
					uploadedFilesNames = append(uploadedFilesNames, randomFileName)
				} else {
					return pathErr, nil
				}
			}
		}
		return nil, uploadedFilesNames
	}
	return errors.New("invalid multipartform"), nil
}

func (this *MultipleUpload) ImgCategoryTargetSizes() map[string][]uint {
	if categorySizes, ok := imageSizes[this.ImageCategory]; ok {
		targetSizes := map[string][]uint{}
		for sizeName, size := range categorySizes {
			for _, s := range this.ImageSizes {
				if s == sizeName {
					targetSizes[s] = size
					break
				}
			}
		}
		return targetSizes
	}
	return nil
}

func (this *MultipleUpload) SetLocalDir(localDir string) {
	localDir = filepath.FromSlash(localDir)
	this.localPath = filepath.Join(baseLocalPath, localDir)
}

func (this *MultipleUpload) LocalPath() string {
	if this.localPath != baseLocalPath {
		return this.localPath
	}
	return ""
}

func (this *MultipleUpload) SetCloudDir(cloudDir string) {
	cloudDir = filepath.FromSlash(cloudDir)
	this.cloudPath = filepath.Join(baseCloudPath, cloudDir)
}

func (this *MultipleUpload) CloudPath() string {
	return this.cloudPath
}
func (this *MultipleUpload) GetCloudFilePath(fileName string, opts ...string) string {
	var sizeName string
	if opts != nil && len(opts) > 0 {
		sizeName = strings.ToLower(strings.TrimSpace(opts[0]))
	}
	if sizeName != "" && sizeName != "original" {
		return filepath.Join(this.cloudPath, sizeName, fileName)
	} else {
		return filepath.Join(this.cloudPath, fileName)
	}
}
func generateRandomFileName(extension string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return strconv.Itoa(int(time.Now().UTC().Unix())) + "-" + hex.EncodeToString(randBytes) + extension
}

func detectContentType(fileData []byte) string {
	if fileData != nil {
		filetype := http.DetectContentType(fileData)
		return filetype
	}
	return ""
}

func isValidFileType(requiredFileTypesRaw string, fileData []byte, fileExtension string) (bool, string, string) {
	isValidExtension := false
	isValidContentType := false
	fileType := detectContentType(fileData)
	fileTypeName := ""
	requiredFileTypesRaw = strings.ToLower(strings.Replace(requiredFileTypesRaw, " ", "", -1))
	requiredFileTypes := strings.Split(requiredFileTypesRaw, "|")
	for _, requiredFileType := range requiredFileTypes {
		switch requiredFileType {
		case "image":
			fileTypeName = "image"
			for _, imageExtension := range imageExtensions {
				if imageExtension == fileExtension {
					isValidExtension = true
					break
				}
			}
			if isValidExtension {
				for _, imageContentType := range imageContentTypes {
					if fileType == imageContentType {
						isValidContentType = true
						break
					}
				}
			}
		case "document":
			fileTypeName = "document"
			for _, documentExtension := range documentExtensions {
				if documentExtension == fileExtension {
					isValidExtension = true
					break
				}
			}
			if isValidExtension {
				for _, documentContentType := range documentContentTypes {
					if fileType == documentContentType {
						isValidContentType = true
						break
					}
				}
			}
		case "svg":
			fileTypeName = "svg"
			for _, svgExtension := range svgExtensions {
				if svgExtension == fileExtension {
					isValidExtension = true
					break
				}
			}
			if isValidExtension {
				for _, svgContentType := range svgContentTypes {
					if fileType == svgContentType {
						isValidContentType = true
						break
					}
				}
			}
		case "pdf":
			fileTypeName = "pdf"
			if fileExtension == ".pdf" {
				isValidExtension = true
			}
			if isValidExtension {
				for _, pdfContentType := range pdfContentTypes {
					if fileType == pdfContentType {
						isValidContentType = true
						break
					}
				}
			}
		}

		if isValidExtension {
			break
		}
	}
	//fmt.Println(isValidContentType, isValidExtension, fileType, fileTypeName)
	return isValidContentType && isValidExtension, fileType, fileTypeName
}

func CreateFolderPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0777); err != nil {
			return false, err
		}
	}
	return true, err
}
