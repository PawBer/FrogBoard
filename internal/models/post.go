package models

import (
	"fmt"
	"html"
	"html/template"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	ID        uint
	BoardID   string
	CreatedAt time.Time
	Content   string
	Files     []FileInfo
	Citations []Citation
	PosterIP  net.IP
}

var PostCitationRegex = regexp.MustCompile("&gt;&gt; ([0-9]+)")
var GreentextRegex = regexp.MustCompile("&gt;.+")

func (p Post) FormatCreationDate() template.HTML {
	return template.HTML(p.CreatedAt.UTC().Format("2006-01-02T15:04:05-0700"))
}

func (p Post) FileCount() int {
	return len(p.Files)
}

func (p Post) FormatedContent() template.HTML {
	citationLink := fmt.Sprintf(`<a data-post="$1" class="post-link text-blue-500" href="/%s/$1/">>> $1</a>`, p.BoardID)
	afterCitations := PostCitationRegex.ReplaceAllString(html.EscapeString(p.Content), citationLink)

	var formatedLines []string
	lines := strings.Split(afterCitations, "\r\n")
	for _, line := range lines {
		var newLine string

		if GreentextRegex.MatchString(line) {
			newLine = `<span class="text-green-600">` + line + `</span><br>`
		} else if line == "" {
			newLine = `<br><br>`
		} else {
			newLine = line + `<br>`
		}

		formatedLines = append(formatedLines, newLine)
	}

	return template.HTML(strings.Join(formatedLines, ""))
}

func GetCitations(boardId string, postId uint, content string) []Citation {
	var citations []Citation

	matches := PostCitationRegex.FindAllStringSubmatch(html.EscapeString(content), -1)

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
