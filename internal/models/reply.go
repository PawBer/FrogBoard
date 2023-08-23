package models

import (
	"gorm.io/gorm"
)

type Reply struct {
	Post
	ThreadID uint
}

type ReplyModel struct {
	DbConn *gorm.DB
}

func (m *ReplyModel) GetRepliesToPost(boardId string, threadId uint) ([]Reply, error) {
	var replies []Reply

	result := m.DbConn.Where("board_id = $1 and thread_id = $2", boardId, threadId).Order("id asc").Find(&replies)
	if err := result.Error; err != nil {
		return nil, err
	}

	return replies, nil
}

func (m *ReplyModel) GetLatestReplies(boardId string, threadId, limit int) ([]Reply, error) {
	var replies []Reply

	subquery := m.DbConn.Limit(limit).Where("board_id = $1 and thread_id = $2", boardId, threadId).Order("id desc")
	result := m.DbConn.Model(&Reply{}).Order("id asc").Find(&replies, subquery)
	if err := result.Error; err != nil {
		return nil, err
	}

	return replies, nil
}

func (m *ReplyModel) Get(boardId string, replyId uint) (*Reply, error) {
	var reply Reply

	result := m.DbConn.Order("id asc").Where("board_id = $1 and id = $2", boardId, replyId).Take(&reply)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &reply, nil
}

func (m *ReplyModel) Insert(boardId string, threadId uint, content string) (uint, error) {
	var board Board
	var reply Reply

	m.DbConn.Transaction(func(tx *gorm.DB) error {
		result := m.DbConn.Select("id, last_post_id").Find(&board)
		if err := result.Error; err != nil {
			m.DbConn.Rollback()
			return err
		}

		reply = Reply{
			Post: Post{
				ID:      board.LastPostID + 1,
				BoardID: boardId,
				Content: content,
			},
			ThreadID: threadId,
		}
		result = m.DbConn.Create(&reply)
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

	return reply.ID, nil
}

func (m *ReplyModel) Update(reply *Reply) error {
	result := m.DbConn.Save(reply)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (m *ReplyModel) Delete(id uint) error {
	result := m.DbConn.Delete(&Reply{}, id)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
