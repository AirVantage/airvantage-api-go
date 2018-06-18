package airvantage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const defaultTimeout = 5 * time.Second

var defaultLogger = log.New(os.Stderr, "AV", log.LstdFlags)

// AirVantage API client using oAuth2
type AirVantage struct {
	client *http.Client
	Debug  bool
	host   string
	log    *log.Logger
}

// NewClient logins to AirVantage an returns a new API client.
func NewClient(host, clientID, clientSecret, login, password string) (*AirVantage, error) {

	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: host + "/api/oauth/token",
			AuthURL:  host + "/api/oauth/auth",
		},
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, http.Client{Timeout: defaultTimeout})

	token, err := conf.PasswordCredentialsToken(ctx, login, password)
	if err != nil {
		return nil, err
	}

	return &AirVantage{client: conf.Client(ctx, token), host: host, log: defaultLogger}, nil
}

// get with smart URL formatting.
func (av *AirVantage) get(format string, a ...interface{}) (*http.Response, error) {
	return av.client.Get(av.URL(format, a...))
}

type apiError struct {
	Error           string
	ErrorParameters []string
}

// parseResponse is the standard way to handle HTTP responses from AirVantage.
// respStruct must be a pointer to a struct where the JSON will be deserialized.
func (av *AirVantage) parseResponse(resp *http.Response, respStruct interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if av.Debug {
			av.log.Printf("Path: %s\nContent: %s\n", resp.Request.URL.Path, string(body))
		}

		if len(body) == 0 {
			return fmt.Errorf("error %d %s", resp.StatusCode, resp.Status)
		}

		apierror := apiError{}
		if err = json.Unmarshal(body, &apierror); err != nil {
			return fmt.Errorf("unable to parse API error: %s", err)
		}

		if apierror.Error != "" {
			return avError(resp.Request.URL.Path, apierror.Error, apierror.ErrorParameters)
		}
	}

	if respStruct == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(respStruct); err != nil {
		return fmt.Errorf("unable to parse API response: %s", err)
	}

	return nil
}

// SetLogger allows you to set a custom logger instead of Go's default.
func (av *AirVantage) SetLogger(logger *log.Logger) {
	av.log = logger
}

// SetTimeout sets the request timeout of the HTTP client.
func (av *AirVantage) SetTimeout(timeout time.Duration) {
	av.client.Timeout = timeout
}

// URL builds a URL with the right host and prefix for API calls.
func (av *AirVantage) URL(format string, a ...interface{}) string {
	return fmt.Sprintf(av.host+"/api/v1"+format, a...)
}
