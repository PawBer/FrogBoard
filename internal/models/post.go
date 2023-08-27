package models

import (
	"fmt"
	"html/template"
	"regexp"
	"time"
)

type Post struct {
	ID        uint
	BoardID   string
	CreatedAt time.Time
	Content   string
	Files     []FileInfo
}

var PostReferenceRegex = regexp.MustCompile(">> ([1-9]+)")

func (p Post) FormatCreationDate() string {
	return p.CreatedAt.UTC().Format("2006-01-02T15:04:05-0700")
}

func (p Post) FormatedContent(boardId string) template.HTML {
	replacement := fmt.Sprintf(`<a data-post="$1" class="post-link text-blue-500" href="/%s/$1/">>> $1</a>`, boardId)

	return template.HTML(PostReferenceRegex.ReplaceAllString(p.Content, replacement))
}
