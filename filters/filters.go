package filters

import (
	"io"
	"strings"
	"unicode"
)

var AvailableFilters map[string]Filterer

// should be thread safe
// should not modify input slice
type Filterer interface {
	Filter([]byte) ([]byte, error)
}

type FilterWriter struct {
	Filters []Filterer
	Writer  io.Writer
}

type WordFilter struct {
	Words       map[string]bool
	Replacement string
}

func (fw FilterWriter) Write(p []byte) (int, error) {
	var err error
	filtered := p

	for _, filter := range fw.Filters {
		filtered, err = filter.Filter(filtered)
		if err != nil {
			return 0, err
		}
	}

	return fw.Writer.Write(filtered)
}

func (f WordFilter) Filter(p []byte) ([]byte, error) {
	result := strings.Builder{}
	result.Grow(len(p))
	accum := strings.Builder{}

	for _, r := range string(p) {
		if !unicode.IsSpace(r) {
			accum.WriteRune(r)
			continue
		}

		if f.Words[accum.String()] {
			result.WriteString(f.Replacement)
		} else {
			result.WriteString(accum.String())
		}
		result.WriteRune(r)
		accum.Reset()
	}

	if accum.Len() != 0 {
		if f.Words[accum.String()] {
			result.WriteString(f.Replacement)
		} else {
			result.WriteString(accum.String())
		}
	}

	return []byte(result.String()), nil
}

func init() {
	AvailableFilters = make(map[string]Filterer)

	bannedWords := make(map[string]bool)

	AvailableFilters["testFilter"] = WordFilter{Words: bannedWords, Replacement: ""}
}
