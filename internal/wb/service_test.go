package wb

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func Test_AddNewWordBubble(t *testing.T) {
	tests := map[string]struct {
		wordbubble  *model.WordBubble
		userId      int64
		repo        WordBubbleRepo
		expectedErr string
	}{
		"valid": {
			userId: 3462,
			wordbubble: &model.WordBubble{
				Text: "hello world",
			},
			repo: &testWordBubbleRepo{},
		},
		"invalid wordbubble text less than min bound": {
			userId: 355,
			wordbubble: &model.WordBubble{
				Text: strings.Repeat(".", util.MinWordBubbleLength-1),
			},
			repo: &testWordBubbleRepo{},
			expectedErr: fmt.Sprintf(
				"wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d",
				util.MinWordBubbleLength, util.MaxWordBubbleLength, util.MinWordBubbleLength-1,
			),
		},
		"invalid wordbubble text greater than max bound": {
			userId: 32,
			wordbubble: &model.WordBubble{
				Text: strings.Repeat(".", util.MaxWordBubbleLength+1),
			},
			repo: &testWordBubbleRepo{},
			expectedErr: fmt.Sprintf(
				"wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d",
				util.MinWordBubbleLength, util.MaxWordBubbleLength, util.MaxWordBubbleLength+1,
			),
		},
		"invalid, database error": {
			userId: 3612,
			wordbubble: &model.WordBubble{
				Text: "hello world again",
			},
			repo: &testWordBubbleRepo{
				err: errors.New("explosion"),
			},
			expectedErr: "explosion",
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewWordBubblesService(util.TestLogger(), tcase.repo)
			err := svc.AddNewWordBubble(tcase.userId, tcase.wordbubble)
			if tcase.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr, err.Error())
			}
		})
	}
}

func Test_RemoveAndReturnLatestWordBubbleForUserId(t *testing.T) {
	tests := map[string]struct {
		userId             int64
		repo               WordBubbleRepo
		expectedWordBubble bool
	}{
		"wordbubble returned": {
			userId: 3462,
			repo: &testWordBubbleRepo{
				wordbubble: &model.WordBubble{},
			},
			expectedWordBubble: true,
		},
		"wordbubble not returned": {
			userId: 3462,
			repo:   &testWordBubbleRepo{},
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewWordBubblesService(util.TestLogger(), tcase.repo)
			wordbubble := svc.RemoveAndReturnLatestWordBubbleForUserId(tcase.userId)
			if tcase.expectedWordBubble {
				assert.NotNil(t, wordbubble)
			} else {
				assert.Nil(t, wordbubble)
			}
		})
	}
}

type testWordBubbleRepo struct {
	err        error
	wordbubble *model.WordBubble
}

func (trepo *testWordBubbleRepo) addNewWordBubble(userId int64, wb *model.WordBubble) error {
	return trepo.err
}

func (trepo *testWordBubbleRepo) removeAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	return trepo.wordbubble
}
