package util

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

func IsImageFile(filename string) bool {
	allowedExts := strings.Split(config.SupportedImageFormats, ",")
	for i, ext := range allowedExts {
		allowedExts[i] = strings.TrimSpace(ext)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(allowedExts, ext)
}

func IsFrontendFile(filename string) bool {
	frontendExts := []string{".html"}
	return slices.Contains(frontendExts, strings.ToLower(filepath.Ext(filename)))
}
