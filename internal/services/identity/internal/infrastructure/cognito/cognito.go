package cognito

import (
	"context"
	"errors"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/auth"
	svErr "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoProvider struct {
	cfg config.CognitoConfig
	svc *cognitoidentityprovider.Client
}

func NewCognitoProvider(cfg config.CognitoConfig) (*CognitoProvider, error) {
	if cfg.Region == "" || cfg.UserPoolID == "" || cfg.ClientID == "" {
		return nil, errors.New("cognito: Region, UserPoolID and ClientID are required")
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("cognito: failed to load AWS config: %w", err)
	}

	return &CognitoProvider{
		cfg: cfg,
		svc: cognitoidentityprovider.NewFromConfig(awsCfg),
	}, nil
}

func (cp *CognitoProvider) ProviderName() string {
	return "cognito"
}

func (cp *CognitoProvider) SignUp(ctx context.Context, email, password string, attrs map[string]string) (*auth.AuthProviderSignUpResp, error) {
	attributes := []types.AttributeType{
		{Name: aws.String("email"), Value: aws.String(email)},
	}
	for k, v := range attrs {
		attributes = append(attributes, types.AttributeType{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}
	input := &cognitoidentityprovider.SignUpInput{
		ClientId:       &cp.cfg.ClientID,
		Username:       &email,
		Password:       &password,
		UserAttributes: attributes,
		SecretHash:     aws.String(cp.secretHash(email)),
	}

	out, err := cp.svc.SignUp(ctx, input)
	if err != nil {
		return nil, svErr.ErrInternal.WithCause(err).WithInternal("CognitoProvider.SignUp")
	}
	return &auth.AuthProviderSignUpResp{
		UserSub:    *out.UserSub,
		IsVerified: out.UserConfirmed,
	}, nil
}

func (cp *CognitoProvider) SignIn(ctx context.Context, email, password string) (*auth.AuthProviderSignInResp, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: &cp.cfg.ClientID,
		AuthParameters: map[string]string{
			"USERNAME":    email,
			"PASSWORD":    password,
			"SECRET_HASH": cp.secretHash(email),
		},
	}

	out, err := cp.svc.InitiateAuth(ctx, input)
	if err != nil {
		var notAuthorized *types.NotAuthorizedException
		var userNotFound *types.UserNotFoundException
		var userNotConfirmed *types.UserNotConfirmedException
		switch {
		case errors.As(err, &notAuthorized):
			return nil, svErr.ErrInvalidCredentials
		case errors.As(err, &userNotFound):
			return nil, svErr.ErrUserNotFound
		case errors.As(err, &userNotConfirmed):
			return nil, svErr.ErrUserNotVerified
		default:
			return nil, svErr.ErrInternal.WithCause(err).WithInternal("CognitoProvider.SignIn")
		}
	}
	fmt.Println("=======out", out)
	tokens := out.AuthenticationResult
	return &auth.AuthProviderSignInResp{
		AccessToken:  aws.ToString(tokens.AccessToken),
		IDToken:      aws.ToString(tokens.IdToken),
		RefreshToken: aws.ToString(tokens.RefreshToken),
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (cp *CognitoProvider) ConfirmAccount(ctx context.Context, email, code string) error {
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         &cp.cfg.ClientID,
		Username:         &email,
		ConfirmationCode: &code,
		SecretHash:       aws.String(cp.secretHash(email)),
	}

	_, err := cp.svc.ConfirmSignUp(ctx, input)
	if err != nil {
		var expiredCode *types.ExpiredCodeException
		var codeMismatch *types.CodeMismatchException
		var notAuthorized *types.NotAuthorizedException
		var userNotFound *types.UserNotFoundException
		switch {
		case errors.As(err, &expiredCode):
			return svErr.ErrExpiredConfirmationCode
		case errors.As(err, &codeMismatch):
			return svErr.ErrInvalidConfirmationCode
		case errors.As(err, &notAuthorized):
			return svErr.ErrInvalidCredentials
		case errors.As(err, &userNotFound):
			return svErr.ErrUserNotFound
		default:
			return svErr.ErrInternal.WithCause(err).WithInternal("CognitoProvider.ConfirmAccount")
		}
	}

	return nil
}
