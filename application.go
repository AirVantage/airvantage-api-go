package airvantage

import (
	"fmt"
)

// An Application descriptor.
type Application struct {
	UID                string                 `json:"uid,omitempty"`
	Name               string                 `json:"name"`
	Revision           string                 `json:"revision"`
	Type               string                 `json:"type,omitempty"`
	Category           string                 `json:"category,omitempty"`
	State              string                 `json:"state,omitempty"`
	Released           AVTime                 `json:"released,omitempty"`
	Published          AVTime                 `json:"published,omitempty"`
	Deprecated         AVTime                 `json:"deprecated,omitempty"`
	IsReference        bool                   `json:"isReference,omitempty"`
	IsPublic           bool                   `json:"isPublic,omitempty"`
	Labels             []string               `json:"labels,omitempty"`
	ApplicationManager string                 `json:"applicationManager,omitempty"`
	Owner              map[string]interface{} `json:"owner,omitempty"` // TODO: real impl.
}

type appResponse struct {
	Count int
	Size  int
	Items []Application
}

// FindAppUID looks for an application using its name and revision,
// checks if it is in the published state, and returns its UID.
func (av *AirVantage) FindAppUID(name, rev string) (string, error) {
	resp, err := av.get("applications", "name", name, "revision", rev, "fields", "uid,state", "size", 2)
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
