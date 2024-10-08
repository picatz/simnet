package http

import (
	"net/http"

	"github.com/picatz/simnet"
)

func NewClient(cfg *simnet.Config) *http.Client {
	return &http.Client{
		Transport: &Transport{
			Dialer: simnet.NewDialer(cfg),
		},
	}
}

func WrapClient(client *http.Client, cfg *simnet.Config) {
	client.Transport = &Transport{
		Underlying: client.Transport.(*http.Transport),
		Dialer:     simnet.NewDialer(cfg),
	}
}
