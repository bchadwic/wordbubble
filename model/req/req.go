// req contains the types for requests
package req

// @Description WordbubbleRequest contains the data sent from a user
type WordbubbleRequest struct {
	Text string `json:"text" example:"Hello world, this is just an example of a wordbubble"`
}

// @Description RefreshTokenRequest contains the token string of a refresh token
type RefreshTokenRequest struct {
	Token string `json:"refresh_token" example:"xxx.yyy.zzz"`
}

// @Description PopUserRequest contains the data to remove and return a wordbubble
type PopUserRequest struct {
	User string `json:"user" example:"ben"`
}

// @Description SignupUserRequest contains the data to signup a new user
type SignupUserRequest struct {
	Username string `json:"username" example:"ben"`
	Email    string `json:"email" example:"benchadwick87@gmail.com"`
	Password string `json:"password" example:"Hello123!"`
}

// @Description LoginUser is the body sent to the /login operation
type LoginUser struct {
	User     string `json:"user" example:"ben"`
	Password string `json:"password" example:"SomePassword_123"`
}
