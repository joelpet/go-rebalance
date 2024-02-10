package avanza

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// Client gives access to Avanza's unofficial backend API.
type Client struct {
	req *req.Client
}

func NewClient() (*Client, error) {
	return &Client{
		req: req.C().
			SetBaseURL("https://www.avanza.se"),
	}, nil
}

func (c Client) Authenticate(creds UserCredentials) error {
	if creds.AuthTimeout < 30 || creds.AuthTimeout > 60*24 {
		return fmt.Errorf("avanza: invalid auth timeout: %d", creds.AuthTimeout)
	}

	var payload authenticatePayload

	err := c.req.Post("/_api/authentication/sessions/usercredentials").
		SetBody(creds).
		Do().
		Into(&payload)

	if err != nil {
		return fmt.Errorf("avanza: post auth req: %v", err)
	}

	// TODO: Indicate whether two-factor authentication is necessary

	return nil
}

// TOTP performs a time based one-time password two factor login.
func (c *Client) TOTP(totp TOTP) error {
	var payload totpPayload

	err := c.req.Post("/_api/authentication/sessions/totp").
		SetBody(totp).
		Do().
		Into(&payload)

	if err != nil {
		return fmt.Errorf("avanza: providing TOTP: %s", err)
	}

	return nil
}

func (c *Client) GetPositions() (*PositionsPayload, error) {
	var payload PositionsPayload

	err := c.req.Get("/_api/position-data/positions").
		Do().
		Into(&payload)

	if err != nil {
		return nil, fmt.Errorf("avanza: getting positions: %s", err)
	}

	return &payload, nil
}

func (c *Client) GetPeriodicSavings() (*PeriodicSavingsPayload, error) {
	var payload PeriodicSavingsPayload

	err := c.req.Get("/_api/periodic-fund-saving/get-periodic-savings").
		Do().
		Into(&payload)

	if err != nil {
		return nil, fmt.Errorf("avanza: getting periodic savings: %s", err)
	}

	return &payload, nil
}

// UserCredentials is used for authenticating with username and password.
type UserCredentials struct {
	// Username is the numeric id used to log in on the web site.
	Username string `json:"username"`
	// Password is the static secret used to log in on the web site.
	Password string `json:"password"`
	// TODO: Verify if this is really needed
	// AuthTimeout describes the maximum number of inactive minutes before the session is expired.
	AuthTimeout uint `json:"maxInactiveMinutes"`
}

// TOTP is used for two-factor authentication.
type TOTP struct {
	Method   string `json:"method"`
	TOTPCode string `json:"totpCode"`
}

type authenticatePayload struct {
	TwoFactorLogin struct {
		// Transaction id, e.g. "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
		TransactionID string `json:"transactionId"`
		// 2FA method, e.g. "TOTP"
		Method string `json:"method"`
	} `json:"twoFactorLogin"`
}

type totpPayload struct {
	// Authentication session, e.g. "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	AuthenticationSession string `json:"authenticationSession"`
	// Push subscription id, e.g. "deadb3333333333333333333333333333333333f"
	PushSubscriptionID string `json:"pushSubscriptionId"`
	// Customer id, e.g. "1111111"
	CustomerID string `json:"customerId"`
	// Registration complete, e.g. true
	RegistrationComplete bool `json:"registrationComplete"`
}

type PositionsPayload struct {
	WithOrderbook    []Position `json:"withOrderbook"`
	WithoutOrderbook []Position `json:"withoutOrderbook"`
	CashPositions    []struct {
		Account struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		TotalBalance struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		} `json:"totalBalance"`
	} `json:"cashPositions"`
}

type Position struct {
	Account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	Instrument struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
	} `json:"instrument"`
	Value struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"value"`
}

type PeriodicSavingsPayload struct {
	PeriodicSavings []struct {
		Account struct {
			// Account id, e.g. 5555555
			AccountID int `json:"accountId"`
			// Account name, e.g. "Avanza Framtid"
			AccountName string `json:"accountName"`
		} `json:"account"`
		AllocationViews []struct {
			Allocation  float32 `json:"allocation"`
			Name        string  `json:"name"`
			OrderbookID int     `json:"orderbookId"`
		} `json:"allocationViews"`
		// Monthly savings id, e.g. "A1^1608186314557^55559"
		MonthlySavingsID string `json:"monthlySavingsId"`
	} `json:"periodicSavings"`
}
