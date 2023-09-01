package models

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/doug-martin/goqu/v9"
)

type FileInfo struct {
	ID          string
	Name        string
	ContentType string
}

type FileInfoModel struct {
	DbConn    *goqu.Database
	FileStore filestorage.FileStore
}

func (fi FileInfo) ContainsImage() bool {
	return strings.Contains(fi.ContentType, "image")
}

func (fiModel *FileInfoModel) GetFilesForPost(boardId string, postId uint) ([]FileInfo, error) {
	var fileInfos []FileInfo
	var fileIds []string

	query, params, _ := fiModel.DbConn.From("post_files").Select("file_id").Where(goqu.Ex{
		"board_id": boardId,
		"post_id":  postId,
	}).ToSQL()

	rows, err := fiModel.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	var fileId string
	for rows.Next() {
		err = rows.Scan(&fileId)
		if err != nil {
			return nil, err
		}
		fileIds = append(fileIds, fileId)
	}
	if len(fileIds) == 0 {
		return nil, sql.ErrNoRows
	}

	query, params, _ = fiModel.DbConn.From("file_infos").Select().Where(goqu.C("id").In(fileIds)).ToSQL()

	rows, err = fiModel.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		fileInfo := FileInfo{}

		err = rows.Scan(&fileInfo.ID, &fileInfo.Name, &fileInfo.ContentType)
		if err != nil {
			return nil, err
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

func (fiModel *FileInfoModel) InsertFile(fileName string, file []byte) (string, error) {
	contentType := http.DetectContentType(file)

	key, err := fiModel.FileStore.AddFile(file)
	if err != nil {
		return "", err
	}

	query, params, _ := fiModel.DbConn.Insert("file_infos").Rows(goqu.Record{"id": key, "file_name": fileName, "content_type": contentType}).ToSQL()

	_, err = fiModel.DbConn.Exec(query+" ON CONFLICT (id) DO NOTHING", params...)
	if err != nil {
		return "", err
	}

	return key, nil
}
