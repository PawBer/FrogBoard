package models

import (
	"github.com/doug-martin/goqu/v9"
)

type Citation struct {
	BoardID string
	PostID  uint
	Cites   uint
}

type CitationModel struct {
	DbConn *goqu.Database
}

func (cm *CitationModel) GetCitationsForPosts(boardId string, posts ...*Post) error {
	if len(posts) == 0 {
		return nil
	}

	var ids []uint

	for _, post := range posts {
		ids = append(ids, post.ID)
	}

	sql, params, _ := goqu.From("citations").Select("board_id", "post_id", "cites").Where(goqu.Ex{
		"board_id": boardId,
		"cites":    ids,
	}).ToSQL()

	rows, err := cm.DbConn.Query(sql, params...)
	if err != nil {
		return err
	}

	var citation Citation
	for rows.Next() {
		err = rows.Scan(&citation.BoardID, &citation.PostID, &citation.Cites)
		if err != nil {
			return err
		}

		for _, post := range posts {
			if citation.Cites == post.ID {
				post.Citations = append(post.Citations, citation)
			}
		}
	}

	return nil
}
