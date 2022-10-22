// model contains the data types that are used internally to the api
package model

import "github.com/golang-jwt/jwt"

type TokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

type User struct {
	Id       int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
