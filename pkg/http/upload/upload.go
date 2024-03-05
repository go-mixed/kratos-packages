package upload

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type uploadFile struct {
	fileName     string
	file         io.ReadSeekCloser
	size         int64
	deletingPath string // 临时文件路径, 用于关闭时删除
}

var _ io.ReadSeekCloser = (*uploadFile)(nil)

func (f *uploadFile) Read(p []byte) (n int, err error) {
	if f.file == nil {
		return 0, errors.New("无法读取文件，文件已关闭")
	}
	return f.file.Read(p)
}

func (f *uploadFile) Seek(offset int64, whence int) (int64, error) {
	if f.file == nil {
		return 0, errors.New("无法修改文件指针，文件已关闭")
	}
	return f.Seek(offset, whence)
}

func (f *uploadFile) Close() error {
	if f.file != nil {
		err := f.file.Close()
		if f.deletingPath != "" { // 删除临时文件
			_ = os.Remove(f.deletingPath)
		}
		return err
	}

	f.file = nil
	return nil
}

func (f *uploadFile) Name() string {
	return f.fileName
}

func (f *uploadFile) Size() int64 {
	return f.size
}

// GetFileFromRequest 从Request中获取上传文件
// 1. 如果是multipart/form-data格式，从FormFile中获取
// 2. 如果是其他格式，从Body中获取，文件名从URL?fieldName=xxx中获取
func GetFileFromRequest(r *http.Request, fieldName string, maxSize int64) (*uploadFile, error) {
	if r.Body == nil || r.ContentLength == 0 {
		return nil, errors.New("上传文件不能为空")
	}

	defer r.Body.Close()

	// Multipart form
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		// 解析上传文件, 最大maxSize
		if err := r.ParseMultipartForm(maxSize); err != nil {
			return nil, errors.New("请正确使用multipart/form-data格式上传文件")
		}

		file, handler, err := r.FormFile(fieldName)

		if err != nil {
			return nil, err
		}

		return &uploadFile{
			fileName: handler.Filename,
			file:     file,
			size:     handler.Size,
		}, nil
	}

	if r.ContentLength > maxSize {
		return nil, fmt.Errorf("上传文件请不要超过%.2fMB", float64(maxSize)/1024./1024.)
	}

	// Body as file
	fileName := r.URL.Query().Get(fieldName)
	file, err := os.CreateTemp(os.TempDir(), "kratos-upload-*")
	if err != nil {
		return nil, err
	}
	contentLength := r.ContentLength
	if contentLength, err = io.Copy(file, r.Body); err != nil {
		return nil, err
	}

	// Seek to the beginning of the file
	_, _ = file.Seek(0, 0)

	return &uploadFile{
		fileName:     fileName,
		file:         file,
		size:         contentLength,
		deletingPath: file.Name(),
	}, nil
}
