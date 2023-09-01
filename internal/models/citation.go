package models

import "github.com/doug-martin/goqu/v9"

type Citation struct {
	BoardID string
	PostID  uint
	Cites   uint
}

type CitationModel struct {
	DbConn *goqu.Database
}

func (cm *CitationModel) GetCitationsForPost(boardId string, postId uint) ([]Citation, error) {
	var citations []Citation

	sql, params, _ := goqu.From("citations").Select("board_id", "post_id").Where(goqu.Ex{
		"board_id": boardId,
		"cites":    postId,
	}).ToSQL()

	rows, err := cm.DbConn.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	var citation Citation
	for rows.Next() {
		err = rows.Scan(&citation.BoardID, &citation.PostID)
		if err != nil {
			return nil, err
		}

		citations = append(citations, citation)
	}
	if len(citations) == 0 {
		return []Citation{}, nil
	}

	return citations, nil
}
