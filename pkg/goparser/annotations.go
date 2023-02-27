package goparser

import (
	"go/ast"
	"strings"
)

const ToolPrefix = "//bfjson:"

type Annotations []string

func parseAnnotations(base Annotations, doc *ast.CommentGroup) Annotations {
	if doc == nil {
		return base
	}

	annotations := append(Annotations{}, base...)
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, ToolPrefix) {
			continue
		}

		annotation := strings.TrimPrefix(comment.Text, ToolPrefix)
		annotations = append(annotations, annotation)
	}

	return annotations
}
