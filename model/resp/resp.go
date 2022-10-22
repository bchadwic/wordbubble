// resp contains the types for responses
package resp

// @Description WordbubbleResponse contains the text returned from the database
type WordbubbleResponse struct {
	Text string `json:"text" example:"Hello world, this is just an example of a wordbubble"`
}

// @Description TokenResponse contains an access token, and an optional refresh token
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"xxx.yyy.zzz"`
	RefreshToken string `json:"refresh_token,omitempty" example:"xxx.yyy.zzz"`
}

// @Description PushResponse contains the success text response from pushing a new wordbubble
type PushResponse struct {
	Text string `json:"text" example:"thank you!"`
}
