package models

import "github.com/doug-martin/goqu/v9"

type Board struct {
	ID         string
	FullName   string
	LastPostID uint
}

type BoardModel struct {
	DbConn *goqu.Database
}

func (m *BoardModel) GetBoards() ([]Board, error) {
	var boards []Board

	sql, params, _ := m.DbConn.From("boards").Select("id", "full_name").ToSQL()
	rows, err := m.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, fullName string

		rows.Scan(&id, &fullName)
		board := Board{
			ID:       id,
			FullName: fullName,
		}

		boards = append(boards, board)
	}

	return boards, nil
}

func (m *BoardModel) Insert(id, name string) error {
	sql, params, _ := goqu.Insert("boards").Rows(
		goqu.Record{"id": id, "full_name": name, "last_post_id": 0},
	).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (m *BoardModel) Delete(id string) error {
	sql, params, _ := goqu.Delete("boards").Where(goqu.Ex{"id": id}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (m *BoardModel) Update(board Board) error {
	sql, params, _ := goqu.Update("boards").Set(goqu.Record{
		"full_name":    board.FullName,
		"last_post_id": board.LastPostID,
	}).Where(goqu.Ex{"id": board.ID}).ToSQL()

	_, err := m.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
