package filestorage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/h2non/bimg"
)

type FSFileStore struct {
	sync.Mutex
	directoryPath string
}

func NewFileSystemStore(path string) *FSFileStore {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(path, 0755)
	}

	return &FSFileStore{
		directoryPath: path,
	}
}

func (fs *FSFileStore) AddFile(file []byte) (string, error) {
	hash := sha1.New()
	hash.Write(file)
	hashSlice := hash.Sum(nil)

	hexString := hex.EncodeToString(hashSlice)

	directoryPath := fmt.Sprintf("%s/%s", fs.directoryPath, hexString[0:2])
	if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(directoryPath, 0755)
	}

	filePath := fmt.Sprintf("%s/%s/%s", fs.directoryPath, hexString[0:2], hexString[2:])

	_, err := os.Stat(filePath)
	if err == nil {
		return hexString, nil
	}

	if err := os.WriteFile(filePath, file, 0755); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(file)
	if strings.Contains(contentType, "image") {
		size, err := bimg.Size(file)
		if err != nil {
			return "", err
		}

		var resizedImage []byte
		if size.Width < 600 {
			resizedImage = file
		} else {
			resizedImage, err = bimg.NewImage(file).Resize(size.Width/3, size.Height/3)
			if err != nil {
				return "", err
			}
		}

		thumbPath := fmt.Sprintf("%s/%s/%s.thumb", fs.directoryPath, hexString[0:2], hexString[2:])
		_, err = os.Stat(thumbPath)
		if err == nil {
			return hexString, nil
		}

		if err := os.WriteFile(thumbPath, resizedImage, 0755); err != nil {
			return "", err
		}
	}

	return hexString, nil
}

func (fs *FSFileStore) GetFile(key string) ([]byte, error) {
	directoryPath := fmt.Sprintf("%s/%s", fs.directoryPath, key[0:2])
	if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	filePath := fmt.Sprintf("%s/%s/%s", fs.directoryPath, key[0:2], key[2:])
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (fs *FSFileStore) GetFileThumbnail(key string) ([]byte, error) {
	directoryPath := fmt.Sprintf("%s/%s", fs.directoryPath, key[0:2])
	if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	filePath := fmt.Sprintf("%s/%s/%s.thumb", fs.directoryPath, key[0:2], key[2:])
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (fs *FSFileStore) DeleteFiles(keys ...string) error {
	for _, key := range keys {
		directoryPath := fmt.Sprintf("%s/%s", fs.directoryPath, key[0:2])
		if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
			return nil
		}

		filePath := fmt.Sprintf("%s/%s/%s", fs.directoryPath, key[0:2], key[2:])
		_, err := os.Stat(filePath)
		if err == nil {
			if err := os.Remove(filePath); err != nil {
				return err
			}
		} else {
			return err
		}

		filePath = fmt.Sprintf("%s/%s/%s.thumb", fs.directoryPath, key[0:2], key[2:])
		_, err = os.Stat(filePath)
		if err == nil {
			if err := os.Remove(filePath); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
