package article

import "regexp"

var CommentPattern = regexp.MustCompile("<!--[^>]*>")

type Comment string

func (t Comment) Render() string {
	return ""
}

func ParseComment(s string) Token {
	return Comment(s)
}
