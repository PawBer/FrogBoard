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
	if len(posts) == 0 {
		return nil
	}

	var ids []uint

	for _, post := range posts {
		ids = append(ids, post.ID)
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
