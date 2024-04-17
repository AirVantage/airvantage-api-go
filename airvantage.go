package airvantage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	defaultTimeout = 5 * time.Second
	// regexp pattern to cleanup json from device/internal/securityinfo core API endpoint response
	javaObjectNamespace = `"com\.sierrawireless\.airvantage\.[^"]*",`
)

var defaultLogger = log.New(os.Stderr, "", log.LstdFlags)

// AirVantage API client using oAuth2
type AirVantage struct {
	client     *http.Client
	CompanyUID string
	Debug      bool
	baseURLv1  *url.URL
	baseURLv2  *url.URL
	log        *log.Logger
}

// NewClient logins to AirVantage an returns a new API client.
func NewClient(host, clientID, clientSecret, login, password string) (*AirVantage, error) {

	scheme := "https"
	if strings.HasPrefix(host, "http://") {
		scheme = "http"
		host = strings.TrimPrefix(host, "http://")
	} else {
		host = strings.TrimPrefix(host, "https://")
	}

	oauthURL := &url.URL{Host: host, Scheme: scheme, Path: "/api/oauth/"}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: oauthURL.ResolveReference(&url.URL{Path: "token"}).String(),
			AuthURL:  oauthURL.ResolveReference(&url.URL{Path: "auth"}).String(),
		},
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, http.Client{Timeout: defaultTimeout})

	token, err := conf.PasswordCredentialsToken(ctx, login, password)
	if err != nil {
		return nil, err
	}

	return &AirVantage{
			client:    conf.Client(ctx, token),
			baseURLv1: &url.URL{Host: host, Scheme: scheme, Path: "/api/v1/"},
			baseURLv2: &url.URL{Host: host, Scheme: scheme, Path: "/api/v2/"},
			log:       defaultLogger,
		},
		nil
}

// get with smart URL formatting (API v1)
func (av *AirVantage) get(format string, a ...interface{}) (*http.Response, error) {
	return av.client.Get(av.URL(format, a...))
}

// get with smart URL formatting (API v2)
func (av *AirVantage) getV2(format string, a ...interface{}) (*http.Response, error) {
	return av.client.Get(av.URLv2(format, a...))
}

// get with query parameters (API v1)
func (av *AirVantage) getWithParams(path string, params url.Values) (*http.Response, error) {
	copy := url.Values{}
	for k := range params {
		copy.Add(k, params.Get(k))
	}
	if av.CompanyUID != "" && !copy.Has("company") {
		copy.Add("company", av.CompanyUID)
	}
	return av.client.Get(av.baseURLv1.ResolveReference(&url.URL{Path: path, RawQuery: copy.Encode()}).String())
}

type apiError struct {
	Error           string
	ErrorParameters []string
}

// parseResponse is the standard way to handle HTTP responses from AirVantage.
// respStruct must be a pointer to a struct where the JSON will be deserialized.
func (av *AirVantage) parseResponse(resp *http.Response, respStruct interface{}) error {
	defer resp.Body.Close()

	if err := av.parseError(resp); err != nil {
		return err
	}

	if respStruct == nil {
		return fmt.Errorf("parsing type not set")
	}

	var payload io.Reader = resp.Body
	if av.Debug {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		av.log.Printf("Path: %s\nContent: %s\n", resp.Request.URL, string(body))
		payload = bytes.NewReader(body)
	}

	if err := json.NewDecoder(payload).Decode(respStruct); err != nil {
		return fmt.Errorf("unable to parse API response: %s", err)
	}

	return nil
}

// parseResponseSystemSecurityInfo is similar to parseResponse
// since the response is Java object serialized we have to remove these references
// respStruct must be a pointer to a struct where the JSON will be deserialized.
func (av *AirVantage) parseResponseSystemSecurityInfo(resp *http.Response, respStruct interface{}) error {
	defer resp.Body.Close()

	if err := av.parseError(resp); err != nil {
		return err
	}

	if respStruct == nil {
		return fmt.Errorf("parsing type not set")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if av.Debug {
		av.log.Printf("Path: %s\nContent: %s\n", resp.Request.URL, string(body))
	}

	// use a regexp to remove the Java object reference from the response
	// it's much easier to do that rather than parsing json into a []interface{}
	reg := regexp.MustCompile(javaObjectNamespace)
	jsonFiltered := reg.ReplaceAllString(string(body), "")

	if err := json.Unmarshal([]byte(jsonFiltered), &respStruct); err != nil {
		return fmt.Errorf("unable to parse API response: %s", err)
	}

	return nil
}

func (av *AirVantage) parseError(resp *http.Response) error {
	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if av.Debug {
			av.log.Printf("Path: %s\nContent: %s\n", resp.Request.URL, string(body))
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

// URL builds a URL with the right host and prefix for API calls (API v1)
func (av *AirVantage) URL(path string, a ...interface{}) string {
	v := url.Values{}

	if av.CompanyUID != "" {
		v.Set("company", av.CompanyUID)
	}

	for i := 0; i < len(a); i += 2 {
		if aStr, ok := a[i+1].(string); ok {
			v.Add(a[i].(string), aStr)
		} else {
			v.Add(a[i].(string), fmt.Sprintf("%v", a[i+1]))
		}
	}

	return av.baseURLv1.ResolveReference(&url.URL{Path: path, RawQuery: v.Encode()}).String()
}

// URLv2 builds a URL with the right host and prefix for API calls (api/v2 prefix)
func (av *AirVantage) URLv2(path string, a ...interface{}) string {
	v := url.Values{}

	if av.CompanyUID != "" {
		v.Set("company", av.CompanyUID)
	}

	for i := 0; i < len(a); i += 2 {
		if aStr, ok := a[i+1].(string); ok {
			v.Add(a[i].(string), aStr)
		} else {
			v.Add(a[i].(string), fmt.Sprintf("%v", a[i+1]))
		}
	}

	return av.baseURLv2.ResolveReference(&url.URL{Path: path, RawQuery: v.Encode()}).String()
}
