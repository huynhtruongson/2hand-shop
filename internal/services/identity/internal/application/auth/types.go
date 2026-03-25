package auth

type SignUpParams struct {
	Name     string
	Email    string
	Gender   string
	Password string
}
type SignUpResult struct {
	UserID     string
	IsVerified bool
}
type SignInResult struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int32
}
type ConfirmAccountParams struct {
	Email            string
	ConfirmationCode string
}
type AuthProviderSignUpResp struct {
	UserSub    string
	IsVerified bool
}
type AuthProviderSignInResp struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int32
}
