package filestorage

type FileStore interface {
	AddFile([]byte) (string, error)
	GetFile(string) ([]byte, error)
	GetFileThumbnail(string) ([]byte, error)
	DeleteFile(string) error
}
