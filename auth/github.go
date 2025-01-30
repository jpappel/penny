package auth

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func register(ctx context.Context) {
	conf := &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}

	verifier := oauth2.GenerateVerifier()
	url := conf.AuthCodeURL("", oauth2.S256ChallengeOption(verifier))

	// TODO: get auth code
	var code string

	tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		slog.ErrorContext(ctx, "Error occured while getting oauth2 token", slog.Any("error", err))
		panic(err)
	}

	client := conf.Client(ctx, tok)
	fmt.Println(url, client)
}
