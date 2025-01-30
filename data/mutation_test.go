package data_test

import (
	"testing"
)

func TestPostComment(t *testing.T) {
	testCases := []CommentsTestCase{
		// TODO: test invalid page
		// TODO: test invalid user
		// TODO: test no parent
		// TODO: test parent
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestDeleteComment(t *testing.T) {
	testCases := []CommentsTestCase{
        // TODO: test no comment
        // TODO: test already deleted comment
        // TODO: test normal deletion
    }
	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestHideComment(t *testing.T) {
	testCases := []CommentsTestCase{}
	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}
