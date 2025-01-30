package data_test

import (
	"context"
	"database/sql"
	"testing"

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
				return p.GetPageCommentsById(ctx, -1)
			}},
		{"SingleComment",
			singleCommentPage,
			nil,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
			}},
		{"NestedCommentChain",
			nestedCommentChainPage,
			nil,
			nestedCommentChain,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
			},
		},
		{"CommentForest",
			commentForestPage,
			nil,
			commentForest,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
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
			singleCommentPage,
			nil,
			singleComment,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "apples")
			}},
		{"NestedCommentChain",
			nestedCommentChainPage,
			nil,
			nestedCommentChain,
			func(p data.PennyDB) (*data.Page, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "peaches")
			},
		},
		{"CommentForest",
			commentForestPage,
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

// TODO:: rewrite

// func TestGetCommentsById(t *testing.T) {
// 	testCases := []CommentsTestCase{
// 		{"NoValidComment",
// 			nil,
// 			sql.ErrNoRows,
// 			singleComment,
// 			func(p data.PennyDB) (*data.Page, error) {
// 				ctx := context.WithValue(context.Background(), "now", MaxInt64)
// 				comment, err := p.GetCommentById(ctx, 100)
// 				if err != nil {
// 					return nil, err
// 				}
// 				page := new(data.Page)
// 				return data.NewPage([]data.Comment{comment}), err
// 			}},
// 		{"ValidComment",
// 			data.NewPage([]data.Comment{{1, 0, "pie", false, false, time.Unix(0, 0), 0}}),
// 			nil,
// 			singleComment,
// 			func(p data.PennyDB) (*data.Page, error) {
// 				ctx := context.WithValue(context.Background(), "now", MaxInt64)
// 				comment, err := p.GetCommentById(ctx, 1)
// 				return data.NewPage([]data.Comment{comment}), err
// 			}},
// 		{"HiddenComment",
// 			data.NewPage([]data.Comment{{1, 0, "pie", true, false, time.Unix(0, 0), 0}}),
// 			nil,
// 			hiddenComment,
// 			func(p data.PennyDB) (*data.Page, error) {
// 				ctx := context.WithValue(context.Background(), "now", MaxInt64)
// 				comment, err := p.GetCommentById(ctx, 1)
// 				return data.NewPage([]data.Comment{comment}), err
// 			},
// 		},
// 		{"DeletedComment",
// 			data.NewPage([]data.Comment{{1, 0, "", false, true, time.Unix(0, 0), 0}}),
// 			nil,
// 			deletedComment,
// 			func(p data.PennyDB) (*data.Page, error) {
// 				ctx := context.WithValue(context.Background(), "now", MaxInt64)
// 				comment, err := p.GetCommentById(ctx, 1)
// 				return data.NewPage([]data.Comment{comment}), err
// 			},
// 		},
// 	}
//
// 	for _, testCase := range testCases {
// 		t.Run(testCase.name, testCase.Test)
// 	}
// }
