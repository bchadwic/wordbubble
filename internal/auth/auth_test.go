package auth

type TestAuthRepo struct {
	err          error
	refreshToken *refreshToken
}

func (trepo *TestAuthRepo) StoreRefreshToken(token *refreshToken) error {
	return trepo.err
}

func (trepo *TestAuthRepo) ValidateRefreshToken(token *refreshToken) error {
	return trepo.err
}

func (trepo *TestAuthRepo) GetLatestRefreshToken(userId int64) *refreshToken {
	return trepo.refreshToken
}
