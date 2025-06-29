package services

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type AuthService struct {
	Provider     *oidc.Provider
	Oauth2Config *oauth2.Config
}

func NewAuthService(provider *oidc.Provider, config *oauth2.Config) *AuthService {
	return &AuthService{
		Provider:     provider,
		Oauth2Config: config,
	}
}

func (s *AuthService) IsConfigured() bool {
	return s.Provider != nil && s.Oauth2Config != nil
}
