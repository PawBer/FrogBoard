package models

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"time"
)

type Post struct {
	ID        uint
	BoardID   string
	CreatedAt time.Time
	Content   string
	Files     []FileInfo
	Citations []Citation
}

var PostReferenceRegex = regexp.MustCompile(">> ([1-9]+)")

func (p Post) FormatCreationDate() string {
	return p.CreatedAt.UTC().Format("2006-01-02T15:04:05-0700")
}

func (p Post) FormatedContent() template.HTML {
	replacement := fmt.Sprintf(`<a data-post="$1" class="post-link text-blue-500" href="/%s/$1/">>> $1</a>`, p.BoardID)

	return template.HTML(PostReferenceRegex.ReplaceAllString(p.Content, replacement))
}

func GetCitations(boardId string, postId uint, content string) []Citation {
	var citations []Citation

	matches := PostReferenceRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		citationId, _ := strconv.ParseInt(match[1], 10, 32)

		citation := Citation{
			BoardID: boardId,
			PostID:  postId,
			Cites:   uint(citationId),
		}

		citations = append(citations, citation)
	}

	return citations
}
