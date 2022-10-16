package resp

type WordBubble struct {
	Text string `json:"text" example:"hello world"`
}

type AuthenticatedResponse struct {
	RefreshToken string `json:"refresh_token" example:"xxx.yyy.zzz"`
	AccessToken  string `json:"access_token" example:"xxx.yyy.zzz"`
}
