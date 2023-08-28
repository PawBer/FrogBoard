package models

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Thread struct {
	Post
	Title string
}

type ThreadModel struct {
	DbConn *goqu.Database
}

func (t Thread) GetType() string {
	return "thread"
}

func (m *ThreadModel) GetLatest(boardId string) ([]Thread, error) {
	var threads []Thread

	sql, params, _ := m.DbConn.From("threads").Select("id", "board_id", "created_at", "content", "title").Where(goqu.Ex{
		"board_id": boardId,
	}).Order(goqu.I("id").Desc()).Limit(15).ToSQL()

	rows, err := m.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id uint
		var boardId, content, title string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &title)
		thread := Thread{
			Post: Post{
				ID:        id,
				BoardID:   boardId,
				CreatedAt: creationTime,
				Content:   content,
			},
			Title: title,
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

func (m *ThreadModel) Get(boardId string, threadId uint) (*Thread, error) {
	var thread Thread

	sql, params, _ := m.DbConn.From("threads").Select("id", "board_id", "created_at", "content", "title").Where(goqu.Ex{
		"board_id": boardId,
		"id":       threadId,
	}).ToSQL()

	row := m.DbConn.QueryRow(sql, params...)

	err := row.Scan(&thread.ID, &thread.BoardID, &thread.CreatedAt, &thread.Content, &thread.Title)
	if err != nil {
		return nil, err
	}

	return &thread, nil
}

func (m *ThreadModel) Insert(boardId, title, content string, files []string) (uint, error) {
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

	sql, params, _ = m.DbConn.Insert("threads").Rows(goqu.Record{
		"id":         board.LastPostID + 1,
		"board_id":   boardId,
		"content":    content,
		"created_at": goqu.V("NOW()"),
		"title":      title,
	}).ToSQL()

	var lastInsertId uint
	err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&lastInsertId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(files) != 0 {
		var rows []goqu.Record

		for _, file := range files {
			row := goqu.Record{
				"board_id": boardId,
				"post_id":  lastInsertId,
				"file_id":  file,
			}

			rows = append(rows, row)
		}

		sql, params, _ = goqu.Insert("post_files").Rows(rows).ToSQL()

		_, err = tx.Exec(sql, params...)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
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

func (m *ThreadModel) Update(thread *Thread) error {
	sql, params, _ := goqu.Update("threads").Set(goqu.Record{
		"title":   thread.Title,
		"content": thread.Content,
	}).Where(goqu.Ex{"id": thread.ID}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (m *ThreadModel) Delete(id uint) error {
	sql, params, _ := goqu.Delete("threads").Where(goqu.Ex{"id": id}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
