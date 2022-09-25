package user

// import (
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"testing"

// 	"github.com/bchadwic/wordbubble/util"
// 	"github.com/golang-jwt/jwt"
// 	"github.com/stretchr/testify/assert"
// )

// func Test_GenerateAccessToken(t *testing.T) {
// 	tests := map[string]struct {
// 		timer  util.Timer
// 		userId int64
// 	}{
// 		"valid": {
// 			timer:  util.Unix(0),
// 			userId: 245,
// 		},
// 	}
// 	for tname, tcase := range tests {
// 		t.Run(tname, func(t *testing.T) {
// 			jwt.TimeFunc = tcase.timer.Now
// 			svc := NewAuthService(util.TestLogger(), nil, tcase.timer, "test signing key")
// 			tokenStr := svc.GenerateAccessToken(tcase.userId)
// 			parts := strings.Split(tokenStr, ".")
// 			assert.Equal(t, 3, len(parts))
// 		})
// 	}
// }
