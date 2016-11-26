package util

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

	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/nehmeroumani/pill.go/clean"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
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

	baseUploadDirPath string
)

func InitUploader(baseUploadDirectryPath string, imgSizes map[string]map[string][]uint) {
	imageSizes = imgSizes
	baseUploadDirPath = baseUploadDirectryPath
}

type MultipleUpload struct {
	uploadDirectoryPath string
	FormData            *multipart.Form
	FilesInputName      string
	FileType            string
	ImageSizes          []string
	ImageCategory       string
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

			isValidFileType, fileType, fileTypeName := isValidFileType(this.FileType, file, fileExtension)

			if !isValidFileType {
				return errors.New("invalid_file_type"), nil
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				clean.Error(err)
				return err, nil
			}
			randomFileName := generateRandomFileName(fileExtension)
			if ok, pathErr := CreateFolderPath(this.uploadDirectoryPath); ok {
				out, err := os.Create(filepath.Join(this.uploadDirectoryPath, randomFileName))
				defer out.Close()
				if err != nil {
					clean.Error(errors.New("Unable to create the file for writing. Check your write access privilege : " + err.Error()))
					return err, nil
				}

				if fileTypeName == "svg" {
					svgMinifyer := Minifyer()
					err = svgMinifyer.Minify("image/svg+xml", out, file)
				} else {
					_, err = io.Copy(out, file)
				}
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
					resizeImg(randomFileName, this.uploadDirectoryPath, this.ImageCategory, this.ImageSizes, file, fileType)
				}
				uploadedFilesNames = append(uploadedFilesNames, randomFileName)
			} else {
				return pathErr, nil
			}
		}
		return nil, uploadedFilesNames
	}
	return errors.New("invalid multipartform"), nil
}

func (this *MultipleUpload) SetUploadDirectoryPath(directoryPath string) {
	this.uploadDirectoryPath = filepath.Join(baseUploadDirPath, directoryPath)
}

func (this *MultipleUpload) UploadDirectoryPath() string {
	return this.uploadDirectoryPath
}

func generateRandomFileName(extension string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return strconv.Itoa(int(time.Now().UTC().Unix())) + "-" + hex.EncodeToString(randBytes) + extension
}

func detectContentType(file multipart.File) string {
	if file != nil {
		buff := make([]byte, 512)
		_, err := file.Read(buff)
		if err != nil {
			clean.Error(err)
			return ""
		}
		filetype := http.DetectContentType(buff)
		return filetype
	}
	return ""
}

func resizeImg(fileName string, upDirPath string, imageCategory string, targetSizes []string, file multipart.File, fileType string) {
	if file != nil && fileType != "" && fileName != "" && upDirPath != "" && imageSizes != nil {
		var img image.Image
		var err error
		img, _, err = image.Decode(file)
		defer file.Close()
		if err != nil {
			clean.Error(err)
			return
		}
		if s, exist := imageSizes[imageCategory]; exist {
			for _, sizeName := range targetSizes {
				if size, ok := s[sizeName]; ok {
					if pathOk, pathErr := CreateFolderPath(filepath.Join(upDirPath, sizeName)); pathOk {
						m := thumbnail(size[0], size[1], img, resize.Lanczos3)
						out, err := os.Create(filepath.Join(upDirPath, sizeName, fileName))
						if err != nil {
							clean.Error(err)
						}
						defer out.Close()
						if size[0] > 0 && size[1] > 0 {
							m, err = cutter.Crop(m, cutter.Config{
								Width:  int(size[0]),
								Height: int(size[1]),
								Mode:   cutter.Centered,
							})
							if err != nil {
								clean.Error(err)
							}
						}
						switch fileType {
						case "image/jpeg", "image/jpg":
							err = jpeg.Encode(out, m, nil)
						case "image/png":
							err = png.Encode(out, m)
						case "image/gif":
							err = gif.Encode(out, m, nil)
						}
					} else {
						clean.Error(pathErr)
						return
					}
				}
			}
		}
	}
}

func isValidFileType(requiredFileTypesRaw string, file multipart.File, fileExtension string) (bool, string, string) {
	isValidExtension := false
	isValidContentType := false
	fileType := detectContentType(file)
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

func thumbnail(minW uint, minH uint, img image.Image, interp resize.InterpolationFunction) image.Image {
	origBounds := img.Bounds()
	origWidth := float64(origBounds.Dx())  //902
	origHeight := float64(origBounds.Dy()) //902
	newWidth, newHeight := origWidth, origHeight

	minHeight := float64(minH) // 80
	minWidth := float64(minW)  // 80

	if minW > 0 && minH > 0 {
		// Return original image if it have same size as constraints
		if minWidth == origWidth && minHeight == origHeight {
			return img
		}

		if origWidth > minWidth && origHeight > minWidth {
			scale := origWidth / minWidth //902 / 80 = 11.275
			origWidth /= scale            // 902 / 11.275 = 80
			origHeight /= scale           // 902 / 11.275 = 80
		}

		if origWidth < minWidth && origHeight < minWidth {
			scale := minWidth / origWidth
			origWidth *= scale
			origHeight *= scale
		}

		if origWidth < minWidth {
			//origWidth -> origHeight
			//minWidth -> minHeight
			newHeight = (origHeight * minWidth) / origWidth
			newWidth = minWidth

			if newHeight < minHeight {
				//origWidth -> origHeight
				//minWidth -> minHeight
				newWidth = minHeight * origWidth / origHeight
				newHeight = minHeight
			}
		} else if origHeight < minHeight { //375 < 400
			//origWidth -> origHeight
			//minWidth -> minHeight
			newWidth = (origWidth * minHeight) / origHeight //500 * 400 / 375 = 533
			newHeight = minHeight                           //400

			if newWidth < minWidth { //533 > 500
				//origWidth -> origHeight
				//minWidth -> minHeight
				newHeight = minWidth * origHeight / origWidth //500 * 375 / 500 = 375
				newWidth = minWidth                           //500
			}
		} else {
			newHeight = origHeight
			newWidth = origWidth
		}
	} else {
		newHeight = float64(minH)
		newWidth = float64(minW)
	}
	return resize.Resize(uint(newWidth), uint(newHeight), img, interp)
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

func Minifyer() *minify.M {
	m := minify.New()
	m.AddFunc("image/svg+xml", svg.Minify)
	return m
}
