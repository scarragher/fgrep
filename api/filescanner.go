package fgrep

import (
	"net/http"
	"strings"
)

// extensions that won't be searched
var (
	validContentTypes = []string{"text/"}
)

// FileScanner Defines a scanner that scans file contents for matches
type FileScanner interface {
	Scan(fileContents []byte, searchContent string) (match bool, ok bool)
}

// TextScanner scans a plain text file
type TextScanner struct {
}

// Scan scans a paint text file for a specific string
func (t *TextScanner) Scan(fileContents []byte, searchContent string) (match bool, ok bool) {
	plainText := string(fileContents)
	segments := strings.Split(plainText, "\n")

	for _, s := range segments {

		if strings.Contains(strings.ToLower(s), strings.ToLower(searchContent)) {
			return true, true
		}
	}

	return false, true
}

// Scan scans a specific file's contents using the file extension to select the correct FileScanner to use
func Scan(searchContent string, fileExt string, fileContents []byte) (match bool, ok bool) {

	if len(fileContents) == 0 {
		return false, false
	}

	contentType := http.DetectContentType(fileContents)

	if !validContentType(contentType) {
		return false, false
	}

	var scanner FileScanner

	switch fileExt {
	case "txt":
		scanner = &TextScanner{}
	default:
		scanner = &TextScanner{}
	}

	match, ok = scanner.Scan(fileContents, searchContent)
	return match, ok
}

func validContentType(contentType string) bool {
	for _, validContentType := range validContentTypes {
		content := contentType[0:len(validContentType)]

		if content == validContentType {
			return true
		}
	}

	return false
}
