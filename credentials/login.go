package credentials

import (
	"os"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	isatty "github.com/mattn/go-isatty"
	"github.com/netlify/open-api/go/models"
	"github.com/skratchdot/open-golang/open"
)

const (
	netlifyAPIScheme = "https"
	netlifyAPIHost   = "api.netlifystg.com"
	netlifyTicketURL = "https://app.netlifystg.com/authorize?response_type=ticket&ticket="
)

var apiSchemes = []string{netlifyAPIScheme}

func login(clientID, host string) (string, error) {
	if !isTTY() {
		// do not try to login when the standard input is not a TTY.
		return "", nil
	}

	client, ctx := newNetlifyAPIClient(noCredentials)
	ticket, err := client.CreateTicket(ctx, clientID)
	if err != nil {
		return "", err
	}

	if err := openAuthUI(ticket); err != nil {
		return "", err
	}

	if !ticket.Authorized {
		a, err := client.WaitUntilTicketAuthorized(ctx, ticket)
		if err != nil {
			return "", err
		}

		ticket = a
	}

	token, err := client.ExchangeTicket(ctx, ticket.ID)
	if err != nil {
		return "", err
	}

	if err := saveAccessToken(token.AccessToken); err != nil {
		return "", err
	}

	if err := tryAccessToken(host, token.AccessToken); err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func openAuthUI(ticket *models.Ticket) error {
	return open.Run(netlifyTicketURL + ticket.ID)
}

func noCredentials(r runtime.ClientRequest, _ strfmt.Registry) error {
	r.SetHeaderParam("User-Agent", "git-credential-netlify")
	return nil
}

func isTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
