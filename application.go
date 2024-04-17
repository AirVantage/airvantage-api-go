package airvantage

import (
	"fmt"
	"io"
)

// An Application descriptor.
type Application struct {
	UID                string         `json:"uid,omitempty"`
	Name               string         `json:"name,omitempty"`
	Revision           string         `json:"revision,omitempty"`
	Type               string         `json:"type,omitempty"`
	Category           string         `json:"category,omitempty"`
	State              string         `json:"state,omitempty"`
	Released           AVTime         `json:"released,omitempty"`
	Published          AVTime         `json:"published,omitempty"`
	Deprecated         AVTime         `json:"deprecated,omitempty"`
	IsReference        bool           `json:"isReference,omitempty"`
	IsPublic           bool           `json:"isPublic,omitempty"`
	Labels             []string       `json:"labels,omitempty"`
	ApplicationManager string         `json:"applicationManager,omitempty"`
	Owner              map[string]any `json:"owner,omitempty"` // TODO: real impl.
}

// FindAppUID looks for an application using its name and revision,
// checks if it is in the published state, and returns its UID.
func (av *AirVantage) FindAppUID(name, rev string) (string, error) {
	resp, err := av.get("applications", "name", name, "revision", rev, "fields", "uid,state", "size", 2)
	if err != nil {
		return "", err
	}

	res := struct{ Items []Application }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}

	found := len(res.Items)

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

// FindAppByTypeRev retrieves an application by type and revision
func (av *AirVantage) FindAppByTypeRev(apptype, apprev string) (*Application, error) {
	resp, err := av.get("applications", "type", apptype, "revision", apprev)
	if err != nil {
		return nil, err
	}

	res := struct{ Items []Application }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	if len(res.Items) == 0 {
		return nil, fmt.Errorf("no application with matching type '%s' & revision '%s'", apptype, apprev)
	}

	return &res.Items[0], nil
}

// ReleaseApplication releases an application
func (av *AirVantage) ReleaseApplication(zipFile io.Reader) (string, error) {

	// why do we need /api/v1 prefix here?
	url := av.URL("/api/v1/operations/applications/release")

	resp, err := av.client.Post(url, "application/zip", zipFile)
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}
