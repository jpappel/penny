package filters

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// TODO: figure out how to make composable
//       it might not make sense to make these composable
type Converter interface {
	Convert([]byte, io.Writer)
}

type MarkdownConverter struct {
	goldmark.Markdown
}

// TODO: add config for extensions
//       * KaTex
//       * mermaid?
// TODO: write extension for adding penny classes to nodes

// create a markdown converter with sane defaults
func NewMarkdownConverter() MarkdownConverter {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	return MarkdownConverter{md}
}
