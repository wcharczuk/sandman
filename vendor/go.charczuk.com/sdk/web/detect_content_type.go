package web

import (
	"net/http"
	"os"
	"path/filepath"
)

// KnownExtenions are known extenions mapped to their content types.
var (
	KnownExtensions = map[string]string{
		".html": "text/html; charset=utf-8",
		".xml":  "text/xml; charset",
		".json": "application/json; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
	}
)

// DetectFileContentType generates the content type of a given file by path.
func DetectFileContentType(path string) (string, error) {
	if contentType, ok := KnownExtensions[filepath.Ext(path)]; ok {
		return contentType, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	header := make([]byte, 512)
	_, err = f.Read(header)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(header), nil
}
