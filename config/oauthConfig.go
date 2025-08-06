package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleConfig(clientId, clientSecret string) *oauth2.Config{
	return &oauth2.Config{
    	RedirectURL: "http://localhost:3000/api/auth/google/callback",
    	ClientID: clientId,
    	ClientSecret: clientSecret,
    	Scopes: []string{
    	    "https://www.googleapis.com/auth/userinfo.email",
    	    "https://www.googleapis.com/auth/userinfo.profile",
			"openid",
    	},
    	Endpoint: google.Endpoint,
	}
}