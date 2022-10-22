package resp

type Wordbubble struct {
	Text string `json:"text" example:"hello world"`
}

type TokenResponse struct {
	RefreshToken string `json:"refresh_token,omitempty" example:"xxx.yyy.zzz"`
	AccessToken  string `json:"access_token" example:"xxx.yyy.zzz"`
}

type PushResponse struct {
	Text string `json:"text" example:"thank you!"`
}
