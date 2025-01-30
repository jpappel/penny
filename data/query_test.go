package data_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jpappel/penny/data"
)


func TestGetPageCommentsById(t *testing.T) {
	testCases := []CommentsTestCase{
		{"MissingPage",
			nil,
			sql.ErrNoRows,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, -1, data.SortPaginate{})
			}},
		{"SingleComment",
			data.NewPage([]data.Comment{{1, 0, "pie", false, false, time.Unix(0, 0), 0}}),
			nil,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1, data.SortPaginate{})
			}},
		{"NestedCommentChain",
			data.NewPage([]data.Comment{
				{1, 0, "cobbler", false, false, time.Unix(0, 0), 1},
				{2, 1, "with", false, false, time.Unix(1, 0), 1},
				{3, 2, "icecream", false, false, time.Unix(2, 0), 0},
			}),
			nil,
			nestedCommentChain,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1, data.SortPaginate{})
			},
		},
		{"CommentForest",
			data.NewPage([]data.Comment{
				{1, 0, "first", false, false, time.Unix(0, 0), 2},
				{2, 0, "second", false, false, time.Unix(1, 0), 1},
				{3, 0, "last", false, false, time.Unix(2, 0), 4},
				{4, 1, "letter", false, false, time.Unix(3, 0), 2},
				{5, 1, "animal", false, false, time.Unix(3, 0), 0},
				{6, 2, "ammendment", false, false, time.Unix(4, 0), 1},
				{7, 6, "of the US constitution is the right to bear arms", false, false, time.Unix(5, 0), 0},
				{8, 4, "of the english alphabet descends from proto-sinatic script", false, false, time.Unix(5, 0), 0},
				{9, 4, "is an inverted bull", false, false, time.Unix(5, 0), 0},
				{10, 3, "christmas", false, false, time.Unix(7, 0), 0},
				{11, 3, "I gave you my heart", false, false, time.Unix(8, 0), 0},
				{12, 3, "but then the very next day", false, false, time.Unix(9, 0), 0},
				{13, 3, "you gave it away", false, false, time.Unix(10, 0), 0},
			}),
			nil,
			commentForest,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1, data.SortPaginate{})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetPageComments(t *testing.T) {
	testCases := []CommentsTestCase{
		{"MissingPage",
			nil,
			sql.ErrNoRows,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "I do not exist")
			}},
		{"SingleComment",
			data.NewPage([]data.Comment{{1, 0, "pie", false, false, time.Unix(0, 0), 0}}),
			nil,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "apples")
			}},
		{"NestedCommentChain",
			data.NewPage([]data.Comment{
				{1, 0, "cobbler", false, false, time.Unix(0, 0), 1},
				{2, 1, "with", false, false, time.Unix(1, 0), 1},
				{3, 2, "icecream", false, false, time.Unix(2, 0), 0},
			}),
			nil,
			nestedCommentChain,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "peaches")
			},
		},
		{"CommentForest",
			data.NewPage([]data.Comment{
				{1, 0, "first", false, false, time.Unix(0, 0), 2},
				{2, 0, "second", false, false, time.Unix(1, 0), 1},
				{3, 0, "last", false, false, time.Unix(2, 0), 4},
				{4, 1, "letter", false, false, time.Unix(3, 0), 2},
				{5, 1, "animal", false, false, time.Unix(3, 0), 0},
				{6, 2, "ammendment", false, false, time.Unix(4, 0), 1},
				{7, 6, "of the US constitution is the right to bear arms", false, false, time.Unix(5, 0), 0},
				{8, 4, "of the english alphabet descends from proto-sinatic script", false, false, time.Unix(5, 0), 0},
				{9, 4, "is an inverted bull", false, false, time.Unix(5, 0), 0},
				{10, 3, "christmas", false, false, time.Unix(7, 0), 0},
				{11, 3, "I gave you my heart", false, false, time.Unix(8, 0), 0},
				{12, 3, "but then the very next day", false, false, time.Unix(9, 0), 0},
				{13, 3, "you gave it away", false, false, time.Unix(10, 0), 0},
			}),
			nil,
			commentForest,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "the")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetCommentsById(t *testing.T) {
	testCases := []CommentsTestCase{
		{"NoValidComment",
			nil,
			sql.ErrNoRows,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 100)
				if err != nil {
					return nil, err
				}
				return data.NewPage([]data.Comment{comment}), err
			}},
		{"ValidComment",
			data.NewPage([]data.Comment{{1, 0, "pie", false, false, time.Unix(0, 0), 0}}),
			nil,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return data.NewPage([]data.Comment{comment}), err
			}},
		{"HiddenComment",
			data.NewPage([]data.Comment{{1, 0, "pie", true, false, time.Unix(0, 0), 0}}),
			nil,
			hiddenComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return data.NewPage([]data.Comment{comment}), err
			},
		},
		{"DeletedComment",
			data.NewPage([]data.Comment{{1, 0, "", false, true, time.Unix(0, 0), 0}}),
			nil,
			deletedComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return data.NewPage([]data.Comment{comment}), err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}
