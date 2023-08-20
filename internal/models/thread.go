package models

import (
	"gorm.io/gorm"
)

type Thread struct {
	Post
	Title string
}

type ThreadModel struct {
	DbConn *gorm.DB
}

func (m *ThreadModel) GetLatest(boardID string) ([]Thread, error) {
	var threads []Thread

	result := m.DbConn.Limit(10).Where("board_id = $1", boardID).Order("id desc").Find(&threads)
	if err := result.Error; err != nil {
		return nil, err
	}

	return threads, nil
}

func (m *ThreadModel) Get(boardId string, threadId uint) (*Thread, error) {
	var thread Thread

	result := m.DbConn.Where("board_id = $1 and id = $2", boardId, threadId).Find(&thread)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &thread, nil
}

func (m *ThreadModel) Insert(boardId, title, content string) (uint, error) {
	var board Board
	var thread Thread

	m.DbConn.Transaction(func(tx *gorm.DB) error {
		result := m.DbConn.Select("id, last_post_id").Find(&board)
		if err := result.Error; err != nil {
			m.DbConn.Rollback()
			return err
		}

		thread = Thread{
			Post: Post{
				ID:      board.LastPostID + 1,
				BoardID: boardId,
				Content: content,
			},
			Title: title,
		}
		result = m.DbConn.Create(&thread)
		if err := result.Error; err != nil {
			m.DbConn.Rollback()
			return err
		}

		result = m.DbConn.Model(&board).Update("last_post_id", board.LastPostID+1)
		if err := result.Error; err != nil {
			m.DbConn.Rollback()
			return err
		}

		return nil
	})

	return thread.ID, nil
}

func (m *ThreadModel) Update(thread *Thread) error {
	result := m.DbConn.Save(thread)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (m *ThreadModel) Delete(id uint) error {
	result := m.DbConn.Delete(&Thread{}, id)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
