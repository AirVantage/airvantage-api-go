package airvantage

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
)

// A System descriptor.
type System struct {
	UID                 string                   `json:"uid,omitempty"`
	Name                string                   `json:"name"`
	Type                string                   `json:"type,omitempty"`
	State               string                   `json:"state,omitempty"` // Deprecated
	LifeCycleState      string                   `json:"lifeCycleState,omitempty"`
	ActivityState       string                   `json:"activityState,omitempty"`
	CommStatus          string                   `json:"comStatus,omitempty"`
	CreationDate        AVTime                   `json:"creationDate,omitempty"`
	ActivationDate      AVTime                   `json:"activationDate,omitempty"`
	LastStateChangeDate AVTime                   `json:"lastStateChangeDate,omitempty"`
	LastCommDate        AVTime                   `json:"lastCommDate,omitempty"`
	SyncStatus          string                   `json:"syncStatus,omitempty"`
	LastSyncDate        AVTime                   `json:"lastSyncDate,omitempty"`
	Labels              []string                 `json:"labels,omitempty"`
	Gateway             Gateway                  `json:"gateway,omitempty"`
	Subscription        map[string]string        `json:"subscription,omitempty"`
	Applications        []Application            `json:"applications,omitempty"`
	Metadata            map[string]string        `json:"metadata,omitempty"`
	Data                map[string]interface{}   `json:"data,omitempty"`
	DataUsage           map[string]interface{}   `json:"dataUsage,omitempty"`
	Offer               map[string]interface{}   `json:"offer,omitempty"`
	Communication       Communication            `json:"communication,omitempty"`
	Heatbeat            map[string]interface{}   `json:"heartbeat,omitempty"`
	StatusReport        map[string]string        `json:"statusReport,omitempty"`
	Reports             []map[string]interface{} `json:"reports,omitempty"`
}

// CreateSystem creates a new System on AirVantage. It returns a new System struct
// with updated information. companyUID is an optional argument to change the company context.
// Required fields in System: name, gateway
func (av *AirVantage) CreateSystem(system *System, companyUID string) (*System, error) {

	url := av.URL("/systems?company=%s", companyUID)
	js, err := json.Marshal(system)
	if err != nil {
		return nil, err
	}

	if av.Debug {
		av.log.Printf("Post %s\n%s\n", url, string(js))
	}

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	sys := &System{}
	if err = av.parseResponse(resp, sys); err != nil {
		return nil, err
	}

	return sys, nil
}

// DeleteSystem deletes a system and optionally its gateway and subscription.
func (av *AirVantage) DeleteSystem(uid string, deleteGateway, deleteSubscription bool) error {

	url := av.URL("/systems/%s?deleteGateway=%v&deleteSubscription=%v", uid, deleteGateway, deleteSubscription)

	if av.Debug {
		av.log.Println("DELETE", url)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := av.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}

	return nil
}

// FindSystems is the generic method to find one or more systems.
// Parameters:
// - criteria is a map of field->value to filter the results
// - fields is a comma-seperated list of fields to return (optional)
// - orderBy is a comma-seperated list of fields to order the results (optional)
// You can limit the number of results (100 by default) by adding a criteria 'size'.
func (av *AirVantage) FindSystems(criteria url.Values, fields, orderBy string) ([]System, error) {
	if fields != "" {
		criteria.Set("fields", fields)
	}
	if orderBy != "" {
		criteria.Set("orderBy", orderBy)
	}

	resp, err := av.get("/systems?%s", criteria.Encode())
	if err != nil {
		return nil, err
	}

	var page struct{ Items []System }
	if err = av.parseResponse(resp, &page); err != nil {
		return nil, err
	}

	return page.Items, nil
}

// FindSystemByName returns the first System owning the given name.
// Parameters:
// - fields: a comma-seperated list of fields to return (optional)
// - companyUID: the company context to use (optional)
func (av *AirVantage) FindSystemByName(name, fields, companyUID string) (*System, error) {
	criteria := url.Values{}
	criteria.Set("name", name)
	criteria.Set("size", "1")
	if companyUID != "" {
		criteria.Set("company", companyUID)
	}

	systems, err := av.FindSystems(criteria, fields, "")
	if err != nil || len(systems) == 0 {
		return nil, err
	}

	return &systems[0], err
}

// FindSystemByUID returns the System owning the given UID.
// Parameters:
// - fields: a comma-seperated list of fields to return (optional)
// - companyUID: the company context to use (optional)
func (av *AirVantage) FindSystemByUID(uid, fields, companyUID string) (*System, error) {
	criteria := url.Values{}
	criteria.Set("uid", uid)
	criteria.Set("size", "1")
	if companyUID != "" {
		criteria.Set("company", companyUID)
	}

	systems, err := av.FindSystems(criteria, fields, "")
	if err != nil || len(systems) == 0 {
		return nil, err
	}

	return &systems[0], err
}

// ImportSystems create a batch of systems, with serial numbers ranging from `from` to `to`.
// The systems will be linked to the application `appID` and set to the READY state.
func (av *AirVantage) ImportSystems(from, to int, password, systemType, appID, tag string) error {
	// Genrate the import CSV
	var bb bytes.Buffer
	csvw := csv.NewWriter(&bb)
	csvw.Write([]string{"NAME", "LABELS", "GATEWAY[SERIAL NUMBER]", "MQTT[password]"})

	for serial := from; serial < to; serial++ {
		serialstr := strconv.Itoa(serial)
		csvw.Write([]string{systemType + serialstr, tag, serialstr, password})
		if err := csvw.Error(); err != nil {
			return fmt.Errorf("CSV Writer: %s", err)
		}
	}
	csvw.Flush()

	// Generate the import JSON
	js := fmt.Sprintf(`{"defaultApplications":["%s"],"defaultState":"READY","defaultType":"%s"}`, appID, systemType)

	// Import request
	url := av.URL("/operations/systems/import")
	var b bytes.Buffer
	multi := multipart.NewWriter(&b)

	// CSV part
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="csv"; filename="file.csv"`)
	header.Set("Content-Type", "text/csv")
	partWriter, _ := multi.CreatePart(header)
	partWriter.Write(bb.Bytes())

	// JSON part
	header = make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="parameters"; filename="parameters.json"`)
	header.Set("Content-Type", "application/json")
	partWriter, _ = multi.CreatePart(header)
	partWriter.Write([]byte(js))

	multi.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return fmt.Errorf("NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())

	resp, err := av.client.Do(req)
	if err != nil {
		return err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return err
	}

	op := NewOperation(res.Operation)

	// Waiting for operation to finish
	if av.Debug {
		av.log.Println("waiting for systems import operation", op.UID)
	}
	if err = av.AwaitOperation(op); err != nil {
		return err
	}

	// Check if all the systems were created.
	if op.Counters.Failure > 0 {
		return fmt.Errorf("failed to create %d systems", op.Counters.Failure)
	}

	return nil
}
