package models

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
)

type Board struct {
	ID         string
	FullName   string
	LastPostID uint
	BumpLimit  uint
}

type BoardModel struct {
	DbConn      *goqu.Database
	ThreadModel *ThreadModel
}

func (m *BoardModel) GetBoards() ([]Board, error) {
	var boards []Board

	sql, params, _ := m.DbConn.From("boards").Select("id", "full_name", "last_post_id", "bump_limit").ToSQL()
	rows, err := m.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	var id, fullName string
	var lastPostId, bumpLimit int
	for rows.Next() {

		rows.Scan(&id, &fullName, &lastPostId, &bumpLimit)
		board := Board{
			ID:         id,
			FullName:   fullName,
			LastPostID: uint(lastPostId),
			BumpLimit:  uint(bumpLimit),
		}

		boards = append(boards, board)
	}

	return boards, nil
}

func (m *BoardModel) Insert(id string, name string, bumpLimit uint) error {
	sql, params, _ := goqu.Insert("boards").Rows(
		goqu.Record{"id": id, "full_name": name, "last_post_id": 0, "bump_limit": bumpLimit},
	).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (m *BoardModel) Delete(id string) error {
	query, params, _ := goqu.Delete("boards").Where(goqu.Ex{"id": id}).ToSQL()

	tx, err := m.DbConn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.From("threads").Select("id").Where(goqu.Ex{
		"board_id": id,
	}).ToSQL()

	rows, err := tx.Query(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	var threadIds []uint
	var threadId uint
	for rows.Next() {
		err = rows.Scan(&threadId)
		if err != nil {
			tx.Rollback()
			return err
		}

		threadIds = append(threadIds, threadId)
	}

	if len(threadIds) == 0 {
		tx.Commit()
		return nil
	}

	query, params, _ = goqu.Delete("threads").Where(goqu.Ex{"board_id": id, "id": threadIds}).ToSQL()

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
		"board_id":  id,
		"thread_id": threadIds,
	}).ToSQL()

	rows, err = tx.Query(query, params...)
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
		"board_id":  id,
		"thread_id": threadIds,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.Delete("post_files").Where(goqu.Ex{
		"board_id": id,
		"post_id":  ids,
	}).ToSQL()

	_, err = tx.Exec(query, params...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query, params, _ = goqu.Delete("citations").Where(goqu.Ex{
		"board_id": id,
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

func (m *BoardModel) Update(board Board) error {
	sql, params, _ := goqu.Update("boards").Set(goqu.Record{
		"full_name":  board.FullName,
		"bump_limit": board.BumpLimit,
	}).Where(goqu.Ex{"id": board.ID}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
