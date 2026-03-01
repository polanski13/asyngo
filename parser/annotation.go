package parser

import (
	"go/ast"
	"strings"
)

type annotation struct {
	Name string
	Args []string
	Raw  string
}

type annotationSet struct {
	annotations []*annotation
	byName      map[string][]*annotation
}

func newAnnotationSet(comments *ast.CommentGroup) *annotationSet {
	as := &annotationSet{
		byName: make(map[string][]*annotation),
	}

	if comments == nil {
		return as
	}

	var current *annotation
	for _, c := range comments.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimLeft(text, " ")

		if strings.HasPrefix(text, "@") {
			ann := parseAnnotationLine(text)
			if ann != nil {
				as.annotations = append(as.annotations, ann)
				key := strings.ToLower(ann.Name)
				as.byName[key] = append(as.byName[key], ann)
				current = ann
			}
		} else if current != nil && strings.TrimSpace(text) != "" {
			current.Raw += " " + strings.TrimSpace(text)
		}
	}

	return as
}

func parseAnnotationLine(line string) *annotation {
	line = strings.TrimPrefix(line, "@")
	if line == "" {
		return nil
	}

	nameEnd := 0
	for nameEnd < len(line) && line[nameEnd] != ' ' && line[nameEnd] != '\t' {
		nameEnd++
	}

	name := line[:nameEnd]
	rest := ""
	if nameEnd < len(line) {
		rest = strings.TrimSpace(line[nameEnd:])
	}

	args := tokenizeArgs(rest)

	return &annotation{
		Name: name,
		Args: args,
		Raw:  rest,
	}
}

func tokenizeArgs(s string) []string {
	if s == "" {
		return nil
	}

	var args []string
	var current strings.Builder
	inQuote := false
	inParen := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"' && !inParen:
			if inQuote {
				inQuote = false
				args = append(args, current.String())
				current.Reset()
			} else {
				inQuote = true
			}
		case ch == '(' && !inQuote:
			inParen = true
			current.WriteByte(ch)
		case ch == ')' && !inQuote:
			inParen = false
			current.WriteByte(ch)
			args = append(args, current.String())
			current.Reset()
		case (ch == ' ' || ch == '\t') && !inQuote && !inParen:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func (as *annotationSet) Get(name string) []*annotation {
	return as.byName[strings.ToLower(name)]
}

func (as *annotationSet) GetOne(name string) *annotation {
	anns := as.Get(name)
	if len(anns) == 0 {
		return nil
	}
	return anns[0]
}

func (as *annotationSet) Has(name string) bool {
	return len(as.Get(name)) > 0
}

func (as *annotationSet) All() []*annotation {
	return as.annotations
}
