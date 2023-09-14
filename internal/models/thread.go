package models

import (
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Thread struct {
	Post
	Title   string
	Replies []*Reply
}

type ThreadModel struct {
	DbConn        *goqu.Database
	FileInfoModel *FileInfoModel
	CitationModel *CitationModel
	ReplyModel    *ReplyModel
}

func (t Thread) GetType() string {
	return "thread"
}

func (m *ThreadModel) GetThreadCount(boardId string) (uint, error) {
	query, params, _ := goqu.From("threads").Select(goqu.COUNT("*")).Where(goqu.Ex{
		"board_id": boardId,
	}).ToSQL()

	var count uint
	err := m.DbConn.QueryRow(query, params...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *ThreadModel) GetLatest(boardId string, pageNumber, itemsPerPage uint) ([]*Thread, error) {
	var threads []*Thread

	query, params, _ := goqu.From("threads").Select("id", "board_id", "created_at", "content", "title", "poster_ip").Where(goqu.Ex{
		"board_id": boardId,
	}).Order(goqu.I("last_bump").Desc()).Limit(itemsPerPage).Offset(pageNumber * itemsPerPage).ToSQL()

	rows, err := m.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id uint
		var boardId, content, title, poster_ip string
		var creationTime time.Time

		rows.Scan(&id, &boardId, &creationTime, &content, &title, &poster_ip)
		thread := &Thread{
			Post: Post{
				ID:        id,
				BoardID:   boardId,
				CreatedAt: creationTime,
				Content:   content,
				PosterIP:  net.ParseIP(poster_ip),
			},
			Title: title,
		}

		threads = append(threads, thread)
	}

	var posts []*Post
	for _, thread := range threads {
		posts = append(posts, &thread.Post)
	}

	err = m.ReplyModel.GetLatestReplies(boardId, 5, threads...)
	if err != nil {
		return nil, err
	}

	err = m.FileInfoModel.GetFilesForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = m.CitationModel.GetCitationsForPosts(boardId, posts...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return threads, nil
}

func (m *ThreadModel) Get(boardId string, threadId uint) (*Thread, error) {
	var thread Thread

	query, params, _ := m.DbConn.From("threads").Select("id", "board_id", "created_at", "content", "title", "poster_ip").Where(goqu.Ex{
		"board_id": boardId,
		"id":       threadId,
	}).ToSQL()

	row := m.DbConn.QueryRow(query, params...)

	var posterIp string
	err := row.Scan(&thread.ID, &thread.BoardID, &thread.CreatedAt, &thread.Content, &thread.Title, &posterIp)
	if err != nil {
		return nil, err
	}

	thread.PosterIP = net.ParseIP(posterIp)

	err = m.FileInfoModel.GetFilesForPosts(boardId, &thread.Post)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = m.CitationModel.GetCitationsForPosts(boardId, &thread.Post)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = m.ReplyModel.GetRepliesToThreads(boardId, &thread)
	if err != nil {
		return nil, err
	}

	return &thread, nil
}

func (m *ThreadModel) Insert(boardId, title, content string, files []FileInfo, posterIp string) (uint, error) {
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
		"last_bump":  goqu.V("NOW()"),
		"post_count": 0,
		"poster_ip":  posterIp,
	}).ToSQL()

	var lastInsertId uint
	err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&lastInsertId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(files) != 0 {
		var records []goqu.Record

		for _, file := range files {
			record := goqu.Record{
				"board_id":  boardId,
				"post_id":   lastInsertId,
				"file_id":   file.ID,
				"file_name": file.Name,
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

func (m *ThreadModel) Delete(boardId string, threadIds ...uint) error {
	query, params, _ := goqu.Delete("threads").Where(goqu.Ex{"board_id": boardId, "id": threadIds}).ToSQL()

	tx, err := m.DbConn.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return sql.ErrNoRows
	}

	query, params, _ = goqu.From("replies").Select("id").Where(goqu.Ex{
		"board_id":  boardId,
		"thread_id": threadIds,
	}).ToSQL()

	rows, err := tx.Query(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	var ids []uint
	ids = append(ids, threadIds...)

	var replyId uint
	for rows.Next() {
		err := rows.Scan(&replyId)
		if err != nil {
			tx.Rollback()
			return err
		}

		ids = append(ids, replyId)
	}

	if len(ids) == 0 {
		tx.Commit()
		return nil
	}

	query, params, _ = goqu.Delete("replies").Where(goqu.Ex{
		"board_id":  boardId,
		"thread_id": threadIds,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.Delete("post_files").Where(goqu.Ex{
		"board_id": boardId,
		"post_id":  ids,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.Delete("citations").Where(goqu.Ex{
		"board_id": boardId,
		"post_id":  ids,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
