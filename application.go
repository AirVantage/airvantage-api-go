package airvantage

import (
	"fmt"
	"net/url"
)

// An Application descriptor.
type Application struct {
	UID                string `json:"uid,omitempty"`
	Name               string
	Revision           string
	Type               string                 `json:",omitempty"`
	Category           string                 `json:",omitempty"`
	State              string                 `json:",omitempty"`
	Released           AVTime                 `json:",omitempty"`
	Published          AVTime                 `json:",omitempty"`
	Deprecated         AVTime                 `json:",omitempty"`
	IsReference        bool                   `json:",omitempty"`
	IsPublic           bool                   `json:",omitempty"`
	Labels             []string               `json:",omitempty"`
	ApplicationManager string                 `json:",omitempty"`
	Owner              map[string]interface{} `json:",omitempty"` // TODO: real impl.
}

type appResponse struct {
	Count int
	Size  int
	Items []Application
}

// FindAppUID looks for an application using its name and revision,
// checks if it is in the published state, and returns its UID.
func (av *AirVantage) FindAppUID(name, rev string) (string, error) {
	resp, err := av.get("/applications?name=%s&revision=%s&fields=uid,state&size=2",
		url.QueryEscape(name), rev)
	if err != nil {
		return "", err
	}

	res := appResponse{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}

	found := res.Size

	if found == 0 {
		return "", fmt.Errorf("no application matching '%s'", name)
	}

	// There must be exactly 1 match
	if found > 1 {
		return "", fmt.Errorf("several applications matching '%s'", name)
	}

	// This only contains uid and state
	app := res.Items[0]

	if app.State != "PUBLISHED" {
		return "", fmt.Errorf("application '%s' is not PUBLISHED", name)
	}

	return app.UID, nil
}
