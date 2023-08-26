package models

import (
	"gorm.io/gorm"
)

type Board struct {
	ID         string
	FullName   string
	LastPostID uint
}

type BoardModel struct {
	DbConn *gorm.DB
}

func (m *BoardModel) GetBoards() ([]Board, error) {
	var boards []Board

	result := m.DbConn.Order("id asc").Find(&boards)
	if err := result.Error; err != nil {
		return nil, err
	}

	return boards, nil
}

func (m *BoardModel) Insert(id, name string) error {
	board := Board{
		ID:         id,
		FullName:   name,
		LastPostID: 0,
	}

	result := m.DbConn.Create(&board)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (m *BoardModel) Delete(id string) error {
	result := m.DbConn.Delete(&Board{}, id)
	if err := result.Error; err != nil {
		return err
	}
	result = m.DbConn.Where("board_id = $1", id).Delete(&Post{})
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (m *BoardModel) Update(board *Board) error {
	result := m.DbConn.Save(board)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
