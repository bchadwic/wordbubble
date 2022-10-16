package wb

import (
	"fmt"
	"strings"
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func Test_AddNewWordBubble(t *testing.T) {
	tests := map[string]struct {
		wordbubble  *model.WordBubble
		userId      int64
		repo        WordBubbleRepo
		expectedErr error
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
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", util.MinWordBubbleLength, util.MaxWordBubbleLength, util.MinWordBubbleLength-1),
			),
		},
		"invalid wordbubble text greater than max bound": {
			userId: 32,
			wordbubble: &model.WordBubble{
				Text: strings.Repeat(".", util.MaxWordBubbleLength+1),
			},
			repo: &testWordBubbleRepo{},
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", util.MinWordBubbleLength, util.MaxWordBubbleLength, util.MaxWordBubbleLength+1),
			),
		},
		"invalid, database error": {
			userId: 3612,
			wordbubble: &model.WordBubble{
				Text: "hello world again",
			},
			repo: &testWordBubbleRepo{
				err: resp.InternalServerError("boom"),
			},
			expectedErr: resp.InternalServerError("boom"),
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewWordBubblesService(cfg.TestConfig(), tcase.repo)
			err := svc.AddNewWordBubble(tcase.userId, tcase.wordbubble)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
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
			svc := NewWordBubblesService(cfg.TestConfig(), tcase.repo)
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
