package wb

import (
	"fmt"
	"strings"
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func Test_AddNewWordbubble(t *testing.T) {
	tests := map[string]struct {
		wordbubble  *req.WordbubbleRequest
		userId      int64
		repo        WordbubbleRepo
		expectedErr error
	}{
		"valid": {
			userId: 3462,
			wordbubble: &req.WordbubbleRequest{
				Text: "hello world",
			},
			repo: &testWordbubbleRepo{},
		},
		"invalid wordbubble text less than min bound": {
			userId: 355,
			wordbubble: &req.WordbubbleRequest{
				Text: strings.Repeat(".", util.MinWordbubbleLength-1),
			},
			repo: &testWordbubbleRepo{},
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", util.MinWordbubbleLength, util.MaxWordbubbleLength, util.MinWordbubbleLength-1),
			),
		},
		"invalid wordbubble text greater than max bound": {
			userId: 32,
			wordbubble: &req.WordbubbleRequest{
				Text: strings.Repeat(".", util.MaxWordbubbleLength+1),
			},
			repo: &testWordbubbleRepo{},
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", util.MinWordbubbleLength, util.MaxWordbubbleLength, util.MaxWordbubbleLength+1),
			),
		},
		"invalid, database error": {
			userId: 3612,
			wordbubble: &req.WordbubbleRequest{
				Text: "hello world again",
			},
			repo: &testWordbubbleRepo{
				err: resp.InternalServerError("boom"),
			},
			expectedErr: resp.InternalServerError("boom"),
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewWordbubblesService(cfg.TestConfig(), tcase.repo)
			err := svc.AddNewWordbubble(tcase.userId, tcase.wordbubble)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_RemoveAndReturnLatestWordbubbleForUserId(t *testing.T) {
	tests := map[string]struct {
		userId             int64
		repo               WordbubbleRepo
		expectedWordbubble bool
	}{
		"wordbubble returned": {
			userId: 3462,
			repo: &testWordbubbleRepo{
				wordbubble: &resp.WordbubbleResponse{},
			},
			expectedWordbubble: true,
		},
		"wordbubble not returned": {
			userId: 3462,
			repo:   &testWordbubbleRepo{},
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewWordbubblesService(cfg.TestConfig(), tcase.repo)
			wordbubble := svc.RemoveAndReturnLatestWordbubbleForUserId(tcase.userId)
			if tcase.expectedWordbubble {
				assert.NotNil(t, wordbubble)
			} else {
				assert.Nil(t, wordbubble)
			}
		})
	}
}

type testWordbubbleRepo struct {
	err        error
	wordbubble *resp.WordbubbleResponse
}

func (trepo *testWordbubbleRepo) addNewWordbubble(userId int64, wb *req.WordbubbleRequest) error {
	return trepo.err
}

func (trepo *testWordbubbleRepo) removeAndReturnLatestWordbubbleForUserId(userId int64) *resp.WordbubbleResponse {
	return trepo.wordbubble
}
