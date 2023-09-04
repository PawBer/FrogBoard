package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Reply struct {
	Post
	ThreadID uint
}

type ReplyModel struct {
	DbConn        *goqu.Database
	FileInfoModel *FileInfoModel
	CitationModel *CitationModel
}

func (t Reply) GetType() string {
	return "reply"
}

func (m *ReplyModel) GetRepliesToThreads(boardId string, threads ...*Thread) error {
	var replies []*Reply

	var ids []uint
	for _, thread := range threads {
		ids = append(ids, thread.ID)
	}

	query, params, _ := m.DbConn.From("replies").Select("id", "board_id", "created_at", "content", "thread_id").Where(goqu.Ex{
		"board_id":  boardId,
		"thread_id": ids,
	}).Order(goqu.I("id").Asc()).ToSQL()

	rows, err := m.DbConn.Query(query, params...)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id, threadId uint
		var boardId, content string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &threadId)
		reply := &Reply{
			Post: Post{
				ID:        id,
				BoardID:   boardId,
				CreatedAt: creationTime,
				Content:   content,
			},
			ThreadID: threadId,
		}

		replies = append(replies, reply)
	}
	if len(replies) == 0 {
		return nil
	}

	var posts []*Post
	for _, thread := range replies {
		posts = append(posts, &thread.Post)
	}

	err = m.FileInfoModel.GetFilesForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	err = m.CitationModel.GetCitationsForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	for _, thread := range threads {
		for _, reply := range replies {
			if reply.ThreadID == thread.ID {
				thread.Replies = append(thread.Replies, reply)
			}
		}
	}

	return nil
}

func (m *ReplyModel) GetLatestReplies(boardId string, limit int, threads ...*Thread) error {
	var replies []*Reply

	var ids []uint
	for _, thread := range threads {
		ids = append(ids, thread.ID)
	}

	subquery := m.DbConn.From("replies").Select("*",
		goqu.ROW_NUMBER().Over(goqu.W().PartitionBy("thread_id").OrderBy(goqu.I("id").Desc())).As("ordering"),
	).Where(
		goqu.Ex{"board_id": boardId, "thread_id": ids},
	)

	query, params, _ := m.DbConn.From(subquery).Select("id", "board_id", "created_at", "content", "thread_id").Where(
		goqu.Ex{"ordering": goqu.Op{"lte": limit}},
	).Order(goqu.I("ordering").Desc()).ToSQL()

	rows, err := m.DbConn.Query(query, params...)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id, threadId uint
		var boardId, content string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &threadId)
		reply := &Reply{
			Post: Post{
				ID:        id,
				BoardID:   boardId,
				CreatedAt: creationTime,
				Content:   content,
			},
			ThreadID: threadId,
		}

		replies = append(replies, reply)
	}

	var posts []*Post
	for _, reply := range replies {
		posts = append(posts, &reply.Post)
	}

	err = m.FileInfoModel.GetFilesForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	err = m.CitationModel.GetCitationsForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	for _, thread := range threads {
		for _, reply := range replies {
			if reply.ThreadID == thread.ID {
				thread.Replies = append(thread.Replies, reply)
			}
		}
	}

	return nil
}

func (m *ReplyModel) Get(boardId string, replyId uint) (*Reply, error) {
	reply := Reply{}

	query, params, _ := m.DbConn.From("replies").Select("id", "board_id", "created_at", "content", "thread_id").Where(goqu.Ex{
		"board_id": boardId,
		"id":       replyId,
	}).ToSQL()

	row := m.DbConn.QueryRow(query, params...)

	err := row.Scan(&reply.ID, &reply.BoardID, &reply.CreatedAt, &reply.Content, &reply.ThreadID)
	if err != nil {
		return nil, err
	}

	err = m.FileInfoModel.GetFilesForPosts(boardId, &reply.Post)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = m.CitationModel.GetCitationsForPosts(boardId, &reply.Post)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &reply, nil
}

func (m *ReplyModel) Insert(boardId string, threadId uint, content string, files []string) (uint, error) {
	var board Board

	tx, err := m.DbConn.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	query, params, _ := m.DbConn.From("boards").Where(goqu.Ex{
		"id": boardId,
	}).Select("id", "last_post_id").ToSQL()

	row := tx.QueryRow(query, params...)
	err = row.Scan(&board.ID, &board.LastPostID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	query, params, _ = m.DbConn.Insert("replies").Rows(goqu.Record{
		"id":         board.LastPostID + 1,
		"board_id":   boardId,
		"content":    content,
		"created_at": goqu.V("NOW()"),
		"thread_id":  threadId,
	}).ToSQL()

	var lastInsertId uint
	err = tx.QueryRow(query+" RETURNING id", params...).Scan(&lastInsertId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(files) != 0 {
		var records []goqu.Record

		for _, file := range files {
			record := goqu.Record{
				"board_id": boardId,
				"post_id":  lastInsertId,
				"file_id":  file,
			}

			records = append(records, record)
		}

		query, params, _ := goqu.Insert("post_files").Rows(records).ToSQL()

		_, err := tx.Exec(query, params...)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	citations := GetCitations(boardId, lastInsertId, content)

	if len(citations) != 0 {
		var records []goqu.Record

		for _, citation := range citations {
			record := goqu.Record{
				"board_id": citation.BoardID,
				"post_id":  citation.PostID,
				"cites":    citation.Cites,
			}

			records = append(records, record)
		}

		query, params, _ := goqu.Insert("citations").Rows(records).ToSQL()

		_, err := tx.Exec(query, params...)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	query, params, _ = goqu.Update("boards").Set(goqu.Record{
		"last_post_id": board.LastPostID + 1,
	}).Where(goqu.Ex{"id": board.ID}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return lastInsertId, nil
}

func (m *ReplyModel) Delete(boardId string, id uint) error {
	sql, params, _ := goqu.Delete("replies").Where(goqu.Ex{"board_id": boardId, "id": id}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
