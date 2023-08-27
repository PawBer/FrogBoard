package models

import (
	"database/sql"

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
