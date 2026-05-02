package model

import (
	"errors"
	"path/filepath"
	"strings"
)

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

func ValidateImageExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageExts[ext] {
		return errors.New("image must be jpg, jpeg, png, or webp")
	}
	return nil
}
