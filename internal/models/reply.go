package models

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Reply struct {
	Post
	ThreadID uint
}

type ReplyModel struct {
	DbConn *goqu.Database
}

func (m *ReplyModel) GetRepliesToPost(boardId string, threadId uint) ([]Reply, error) {
	var replies []Reply

	sql, params, _ := m.DbConn.From("replies").Select("id", "board_id", "created_at", "content", "thread_id").Where(goqu.Ex{
		"board_id":  boardId,
		"thread_id": threadId,
	}).Order(goqu.I("id").Asc()).ToSQL()

	rows, err := m.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, threadId uint
		var boardId, content string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &threadId)
		reply := Reply{
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

	return replies, nil
}

func (m *ReplyModel) GetLatestReplies(boardId string, threadId, limit int) ([]Reply, error) {
	var replies []Reply

	subquery := m.DbConn.From("replies").Select("id", "board_id", "created_at", "content", "thread_id").Where(goqu.Ex{
		"board_id":  boardId,
		"thread_id": threadId,
	}).Order(goqu.I("id").Desc()).Limit(5)

	sql, params, _ := m.DbConn.From(subquery).Order(goqu.I("id").Asc()).ToSQL()

	rows, err := m.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, threadId uint
		var boardId, content string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &threadId)
		reply := Reply{
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

	return replies, nil
}

func (m *ReplyModel) Get(boardId string, replyId uint) (*Reply, error) {
	var reply Reply

	sql, params, _ := m.DbConn.From("replies").Select("id", "board_id", "created_at", "content", "thread_id").Where(goqu.Ex{
		"board_id": boardId,
		"id":       replyId,
	}).ToSQL()

	row := m.DbConn.QueryRow(sql, params...)

	err := row.Scan(&reply.ID, &reply.BoardID, &reply.CreatedAt, &reply.Content, &reply.ThreadID)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

func (m *ReplyModel) Insert(boardId string, threadId uint, content string) (uint, error) {
	var board Board

	tx, err := m.DbConn.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	sql, params, _ := m.DbConn.From("boards").Where(goqu.Ex{
		"id": boardId,
	}).Select("id", "last_post_id").ToSQL()

	row := tx.QueryRow(sql, params...)
	err = row.Scan(&board.ID, &board.LastPostID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	sql, params, _ = m.DbConn.Insert("replies").Rows(goqu.Record{
		"id":         board.LastPostID + 1,
		"board_id":   boardId,
		"content":    content,
		"created_at": goqu.V("NOW()"),
		"thread_id":  threadId,
	}).ToSQL()

	var lastInsertId uint
	err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&lastInsertId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	sql, params, _ = goqu.Update("boards").Set(goqu.Record{
		"last_post_id": board.LastPostID + 1,
	}).Where(goqu.Ex{"id": board.ID}).ToSQL()

	_, err = tx.Exec(sql, params...)
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

func (m *ReplyModel) Update(reply *Reply) error {
	sql, params, _ := goqu.Update("replies").Set(goqu.Record{
		"content": reply.Content,
	}).Where(goqu.Ex{"id": reply.ID}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (m *ReplyModel) Delete(id uint) error {
	sql, params, _ := goqu.Delete("replies").Where(goqu.Ex{"id": id}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
