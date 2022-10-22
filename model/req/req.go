package req

type Wordbubble struct {
	Text string `json:"text" example:"Hello world, this is just an example of a wordbubble"`
}

type RefreshToken struct {
	Token string `json:"refresh_token"`
}

// @Description PopUser is the param sent to the /pop operation
type PopUser struct {
	User string `json:"user" example:"ben"`
}

// @Description SignupUser is the body sent to the /signup operation
type SignupUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Description LoginUser is the body sent to the /login operation
type LoginUser struct {
	User     string `json:"user" example:"ben"`
	Password string `json:"password" example:"SomePassword_123"`
}
