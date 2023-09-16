package models

import (
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

func (fiModel *FileInfoModel) GetFilesForPosts(boardId string, posts ...*Post) error {
	var ids []uint

	for _, post := range posts {
		ids = append(ids, post.ID)
	}

	if len(ids) == 0 {
		return nil
	}

	query, params, _ := fiModel.DbConn.From("post_files").Select("post_id", "file_id", "file_name", "content_type").Where(goqu.Ex{
		"board_id": boardId,
		"post_id":  ids,
	}).LeftJoin(
		goqu.T("file_infos"),
		goqu.On(goqu.Ex{"post_files.file_id": goqu.I("file_infos.id")}),
	).ToSQL()

	rows, err := fiModel.DbConn.Query(query, params...)
	if err != nil {
		return err
	}

	var postId uint
	var fileId, fileName, contentType string
	for rows.Next() {
		err = rows.Scan(&postId, &fileId, &fileName, &contentType)
		if err != nil {
			return err
		}

		for _, post := range posts {
			if postId == post.ID {
				post.Files = append(post.Files, FileInfo{ID: fileId, Name: fileName, ContentType: contentType})
			}
		}
	}

	return nil
}

func (fiModel *FileInfoModel) GetLatestFiles() ([]map[string]interface{}, error) {
	fileInfos := []map[string]interface{}{}

	query, params, _ := fiModel.DbConn.From("post_files").Select("post_id", "file_id", "file_name", "board_id").Where(goqu.Ex{}).LeftJoin(
		goqu.T("file_infos"),
		goqu.On(goqu.Ex{"post_files.file_id": goqu.I("file_infos.id")}),
	).Order(goqu.I("post_files.id").Desc()).Limit(15).ToSQL()

	rows, err := fiModel.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	var postId uint
	var fileId, fileName, boardId string
	for rows.Next() {
		err = rows.Scan(&postId, &fileId, &fileName, &boardId)
		if err != nil {
			return nil, err
		}

		fileInfo := map[string]interface{}{}

		fileInfo["BoardID"] = boardId
		fileInfo["PostID"] = postId
		fileInfo["FileID"] = fileId
		fileInfo["Filename"] = postId

		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

func (fiModel *FileInfoModel) InsertFile(fileName string, file []byte) (FileInfo, error) {
	contentType := http.DetectContentType(file)

	key, err := fiModel.FileStore.AddFile(file)
	if err != nil {
		return FileInfo{}, err
	}

	query, params, _ := fiModel.DbConn.Insert("file_infos").Rows(goqu.Record{"id": key, "content_type": contentType}).ToSQL()

	_, err = fiModel.DbConn.Exec(query+" ON CONFLICT (id) DO NOTHING", params...)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{ID: key, Name: fileName, ContentType: contentType}, nil
}

func (fiModel *FileInfoModel) Delete(fileId string) error {
	query, params, _ := goqu.Delete("file_infos").Where(goqu.Ex{
		"id": fileId,
	}).ToSQL()

	tx, err := fiModel.DbConn.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.Delete("post_files").Where(goqu.Ex{
		"file_id": fileId,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = fiModel.FileStore.DeleteFiles(fileId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (fiModel *FileInfoModel) DeleteOrphanedFiles() error {
	query, params, _ := goqu.From("post_files").Select("file_id").ToSQL()

	tx, err := fiModel.DbConn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := tx.Query(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	fileIdsMap := map[string]bool{}

	var fileId string
	for rows.Next() {
		err := rows.Scan(&fileId)
		if err != nil {
			tx.Rollback()
			return err
		}

		fileIdsMap[fileId] = true
	}

	var fileIds []string
	for id := range fileIdsMap {
		fileIds = append(fileIds, id)
	}

	query, params, _ = goqu.Delete("file_infos").Where(goqu.Ex{
		"id": goqu.Op{"notIn": fileIds},
	}).ToSQL()

	rows, err = tx.Query(query+" RETURNING id", params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	var deletedFileIds []string

	for rows.Next() {
		err := rows.Scan(&fileId)
		if err != nil {
			tx.Rollback()
			return err
		}

		deletedFileIds = append(deletedFileIds, fileId)
	}

	err = fiModel.FileStore.DeleteFiles(deletedFileIds...)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
