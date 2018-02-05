package uploader

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	svgContentTypes = []string{"image/svg+xml", "text/xml", "text/xml; charset=utf-8", "text/plain; charset=utf-8"}

	baseLocalUploadPath, baseCloudUploadPath, baseLocalUploadUrlPath, baseCloudUploadUrlPath string
	uploadToCloud                                                                            bool
)

func Init(BaseUploadPath string, BaseUploadUrlPath string, UploadToCloud bool, imgSizes map[string]map[string][]uint) {
	imageSizes = imgSizes
	if !UploadToCloud {
		baseLocalUploadPath = filepath.FromSlash(BaseUploadPath)
		baseLocalUploadUrlPath = BaseUploadUrlPath
	} else {
		baseCloudUploadUrlPath = BaseUploadUrlPath
		baseCloudUploadPath = filepath.FromSlash(BaseUploadPath)
	}
	uploadToCloud = UploadToCloud
}

type MultipleUpload struct {
	FormData           *multipart.Form
	FilesInputName     string
	FileType           string
	ImageSizes         []string
	ImageCategory      string
	localUploadPath    string
	cloudUploadPath    string
	localUploadUrlPath string
	cloudUploadUrlPath string
	WithCrop           bool
	BackgroundColor    *color.RGBA
}

func (this *MultipleUpload) Upload() (error, []string) {
	if this.FormData != nil {
		uploadedFilesNames := []string{}
		errCh := make(chan error, 1)
		finished := make(chan bool, 1)

		files := this.FormData.File[this.FilesInputName] // grab the filenames
		var wg sync.WaitGroup
		for _, file := range files {
			this.UploadOneFile(file, &uploadedFilesNames, errCh, &wg)
		}
		go func() {
			wg.Wait()
			close(finished)
		}()
		select {
		case <-finished:
			return nil, uploadedFilesNames
		case err := <-errCh:
			if err != nil {
				return err, nil
			}
		}
	}
	return errors.New("invalid multipartform"), nil
}

func (this *MultipleUpload) UploadOneFile(fh *multipart.FileHeader, uploadedFilesNames *[]string, errCh chan error, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		file, err := fh.Open()
		defer file.Close()

		if err != nil {
			clean.Error(err)
			errCh <- err
			return
		}
		fileExtension := filepath.Ext(fh.Filename)
		fileExtension = strings.ToLower(fileExtension)

		fileData := make([]byte, 512)
		_, err = file.Read(fileData)
		if err != nil {
			clean.Error(err)
			errCh <- err
			return
		}

		isValidFileType, fileType, fileTypeName := isValidFileType(this.FileType, fileData, fileExtension)

		if !isValidFileType {
			err = errors.New("invalid_file_type")
			errCh <- err
			return
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			clean.Error(err)
			errCh <- err
			return
		}
		randomFileName := generateRandomFileName(fileExtension)
		if uploadToCloud {
			if err = UploadToCloud(gcs.GetClient(), file, this.PathOfFile(randomFileName)); err == nil {
				if fileTypeName == "image" && this.ImageSizes != nil {
					_, err = file.Seek(0, 0)
					if err != nil {
						clean.Error(err)
						errCh <- err
						return
					} else {
						var resizedImages map[string]*izero.Img
						if this.WithCrop {
							resizedImages, err = izero.ResizeImgWithCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes())
						} else {
							resizedImages, err = izero.ResizeImgWithoutCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.BackgroundColor)
						}
						if err != nil {
							clean.Error(err)
						} else {
							var wg sync.WaitGroup
							for sizeName, resizedImage := range resizedImages {
								GoUploadToCloud(gcs.GetClient(), resizedImage.ToReader(), this.PathOfFile(randomFileName, sizeName), &wg)
							}
							wg.Wait()
						}
					}
				}
				*uploadedFilesNames = append(*uploadedFilesNames, randomFileName)
			} else {
				clean.Error(err)
				errCh <- err
				return
			}
		} else {
			if ok, pathErr := CreateFolderPath(this.localUploadPath); ok {
				out, err := os.Create(filepath.Join(this.localUploadPath, randomFileName))
				defer out.Close()
				if err != nil {
					err = errors.New("Unable to create the file for writing. Check your write access privilege : " + err.Error())
					clean.Error(err)
					errCh <- err
					return
				}
				_, err = io.Copy(out, file)

				if err != nil {
					clean.Error(err)
					errCh <- err
					return
				}
				_, err = file.Seek(0, 0)
				if err != nil {
					clean.Error(err)
					errCh <- err
					return
				}
				if fileTypeName == "image" && this.ImageSizes != nil {
					if this.WithCrop {
						izero.ResizeImgWithCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.LocalUploadPath())
					} else {
						izero.ResizeImgWithoutCroping(file, randomFileName, fileType, this.ImgCategoryTargetSizes(), this.LocalUploadPath(), this.BackgroundColor)
					}
				}
				*uploadedFilesNames = append(*uploadedFilesNames, randomFileName)
			} else {
				errCh <- pathErr
				return
			}
		}
	}()
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

func (this *MultipleUpload) SetLocalUploadDir(localDir string) {
	localDir = filepath.FromSlash(localDir)
	this.localUploadPath = filepath.Join(baseLocalUploadPath, localDir)
}

func (this *MultipleUpload) LocalUploadPath() string {
	return this.localUploadPath
}

func (this *MultipleUpload) SetCloudUploadDir(cloudDir string) {
	cloudDir = filepath.FromSlash(cloudDir)
	this.cloudUploadPath = filepath.Join(baseCloudUploadPath, cloudDir)
}

func (this *MultipleUpload) CloudUploadPath() string {
	return this.cloudUploadPath
}

func (this *MultipleUpload) SetUploadDir(dir string) {
	if uploadToCloud {
		this.SetCloudUploadDir(dir)
		if baseCloudUploadUrlPath != "" {
			this.cloudUploadUrlPath = baseCloudUploadUrlPath + "/" + strings.Replace(dir, `\`, "/", -1)
		} else {
			this.cloudUploadUrlPath = strings.Replace(dir, `\`, "/", -1)
		}
	} else {
		this.SetLocalUploadDir(dir)
		this.localUploadUrlPath = baseLocalUploadUrlPath + "/" + strings.Replace(dir, `\`, "/", -1)
	}
}

func (this *MultipleUpload) UploadPath() string {
	if uploadToCloud {
		return this.cloudUploadPath
	} else {
		return this.localUploadPath
	}
}

func (this *MultipleUpload) UploadUrlPath() string {
	if uploadToCloud {
		return this.cloudUploadUrlPath
	} else {
		return this.localUploadUrlPath
	}
}
func (this *MultipleUpload) UrlOfFile(fileName string, opts ...string) string {
	var sizeName string
	if opts != nil && len(opts) > 0 {
		sizeName = strings.ToLower(strings.TrimSpace(opts[0]))
	}
	if uploadToCloud {
		if sizeName != "" && sizeName != "original" {
			return this.cloudUploadUrlPath + "/" + sizeName + "/" + fileName
		} else {
			return this.cloudUploadUrlPath + "/" + fileName
		}
	} else {
		if sizeName != "" && sizeName != "original" {
			return this.localUploadUrlPath + "/" + sizeName + "/" + fileName
		} else {
			return this.localUploadUrlPath + "/" + fileName
		}
	}
}
func (this *MultipleUpload) PathOfFile(fileName string, opts ...string) string {
	var sizeName string
	if opts != nil && len(opts) > 0 {
		sizeName = strings.ToLower(strings.TrimSpace(opts[0]))
	}
	if uploadToCloud {
		if sizeName != "" && sizeName != "original" {
			return filepath.Join(this.cloudUploadPath, sizeName, fileName)
		} else {
			return filepath.Join(this.cloudUploadPath, fileName)
		}
	} else {
		if sizeName != "" && sizeName != "original" {
			return filepath.Join(this.localUploadPath, sizeName, fileName)
		} else {
			return filepath.Join(this.localUploadPath, fileName)
		}
	}
}
func (this *MultipleUpload) AttachmentFileURI(fileName string, opts ...string) string {
	if uploadToCloud {
		return this.UrlOfFile(fileName, opts...)
	} else {
		return this.PathOfFile(fileName, opts...)
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
