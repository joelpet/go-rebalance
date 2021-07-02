package avanza

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

// Client gives access to Avanza's unofficial backend API.
type Client struct {
	httpClient     http.Client
	baseURL        string
	xSecurityToken string
	authSession    string
}

func NewClient() (*Client, error) {
	// TODO: Try without cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("avanza: creating cookie jar: %s", err)
	}
	client := &Client{
		baseURL:    "https://www.avanza.se",
		httpClient: http.Client{Jar: jar},
	}
	return client, nil
}

const ctJSON string = "application/json;charset=utf-8"

func (c Client) Authenticate(creds UserCredentials) error {
	url := c.baseURL + "/_api/authentication/sessions/usercredentials"

	if creds.AuthTimeout < 30 || creds.AuthTimeout > 60*24 {
		return fmt.Errorf("avanza: invalid auth timeout: %d", creds.AuthTimeout)
	}

	reqBody, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("avanza: marshalling login user credentials: %s", err)
	}

	resp, err := c.httpClient.Post(url, ctJSON, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("avanza: authenticating: %s", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("avanza: user credentials authentication: unexpected response status: %s", resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("avanza: reading authentication response body: %s", err)
	}

	var payload authenticatePayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return fmt.Errorf("avanza: unmarshalling authentication payload: %s", err)
	}

	// TODO: Indicate whether two-factor authentication is necessary

	return nil
}

// TOTP performs a time based one-time password two factor login.
func (c *Client) TOTP(totp TOTP) error {
	url := c.baseURL + "/_api/authentication/sessions/totp"

	reqBody, err := json.Marshal(totp)
	if err != nil {
		return fmt.Errorf("avanza: marshalling TOTP request body: %s", err)
	}

	resp, err := c.httpClient.Post(url, ctJSON, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("avanza: providing TOTP: %s", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("avanza: TOTP authentication: unexpected response status: %s", resp.Status)
	}

	if c.xSecurityToken = resp.Header.Get("X-SecurityToken"); c.xSecurityToken == "" {
		return errors.New("avanza: TOTP authentication did not yield the expected security token")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("avanza: reading totp response body: %s", err)
	}

	var payload totpPayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return fmt.Errorf("avanza: unmarshalling totp payload: %s", err)
	}

	if c.authSession = payload.AuthenticationSession; c.authSession == "" {
		return errors.New("avanza: TOTP authentication did not yield the expected authentication session")
	}

	return nil
}

func (c *Client) newJSONReq(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ctJSON)

	if c.authSession != "" {
		req.Header.Set("X-AuthenticationSession", c.authSession)
	}

	if c.xSecurityToken != "" {
		req.Header.Set("X-SecurityToken", c.xSecurityToken)
	}

	return req, nil
}

func (c *Client) GetPositions() (*PositionsPayload, error) {
	url := c.baseURL + "/_mobile/account/positions"

	req, err := c.newJSONReq(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("avanza: creating get positions request: %s", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("avanza: getting positions: %s", err)
	} else if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("avanza: getting positions: unexpected response status: %s", resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("avanza: reading positions response body: %s", err)
	}

	var payload PositionsPayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("avanza: unmarshalling positions payload: %s", err)
	}

	return &payload, nil
}

func (c *Client) GetPeriodicSavings() (*PeriodicSavingsPayload, error) {
	url := c.baseURL + "/_cqbe/fund/periodic-saving/get-periodic-savings"

	req, err := c.newJSONReq(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("avanza: creating get periodic savings request: %s", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("avanza: getting periodic savings: %s", err)
	} else if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("avanza: getting periodic savings: unexpected response status: %s", resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("avanza: reading periodic savings response body: %s", err)
	}

	var payload PeriodicSavingsPayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("avanza: unmarshalling periodic savings payload: %s", err)
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
	InstrumentPositions []struct {
		Positions []struct {
			// Name, e.g. "Avanza Zero"
			Name string `json:"name"`

			// Tradable, e.g. true
			Tradable bool `json:"tradable"`

			// Orderbook id, e.g. "143369"
			OrderbookID string `json:"orderbookId"`

			// Currency, e.g. "SEK"
			Currency string `json:"currency"`

			// Value, e.g. 917.09
			Value float32 `json:"value"`

			// Account id, e.g. "1111111",
			AccountID string `json:"accountId"`

			// Account name, e.g. "Avanza Framtid"
			AccountName string `json:"accountName"`
		} `json:"positions"`
		InstrumentType string `json:"instrumentType"`
	} `json:"instrumentPositions"`
}

type PeriodicSavingsPayload struct {
	PeriodicSavings []struct {
		AccountID       int `json:"accountId"`
		AllocationViews []struct {
			Allocation  float32 `json:"allocation"`
			Name        string  `json:"name"`
			OrderbookID int     `json:"orderbookId"`
		} `json:"allocationViews"`

		AccountInfo struct {
			AccountIdentifier struct {
				// Account id, e.g. "5555555"
				ID string `json:"id"`
			} `json:"accountIdentifier"`
			// Account name, e.g. "Avanza Framtid"
			AccountName string `json:"accountName"`
		} `json:"accountInfo"`
		// Monthly savings id, e.g. "A1^1608186314557^55559"
		MonthlySavingsID string `json:"monthlySavingsId"`
	} `json:"periodicSavings"`
}
