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
	"time"
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

// A Datapoint retrieved from a System.
type Datapoint struct {
	ts AVTime
	v  interface{}
}

// DataAggregate is used to retrieved data from many devices.
// The first string is the systemUID, the second is the data name.
type DataAggregate map[string]map[string][]Datapoint

// ApplyTemplateByUID applies the settings of a given template on a list of systems.
func (av *AirVantage) ApplyTemplateByUID(templateName string, systemUIDs []string) (*Operation, error) {

	reqMsg := struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Template string `json:"templateName"`
	}{Template: templateName}
	reqMsg.Systems.UIDs = systemUIDs

	js, err := json.Marshal(&reqMsg)
	if err != nil {
		return nil, err
	}

	url := av.URL("/operations/systems/settings")

	if av.Debug {
		av.log.Printf("POST %s\n%s\n", url, string(js))
	}

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	if err = av.parseResponse(resp, op); err != nil {
		return nil, err
	}

	return op, nil
}

// ApplyTemplateByLabels applies a template on all the systems with given labels.
func (av *AirVantage) ApplyTemplateByLabels(templateName string, labels []string) (*Operation, error) {

	reqMsg := struct {
		Systems struct {
			Labels []string `json:"labels"`
		} `json:"systems"`
		Template string `json:"templateName"`
	}{Template: templateName}
	reqMsg.Systems.Labels = labels

	js, err := json.Marshal(&reqMsg)
	if err != nil {
		return nil, err
	}

	url := av.URL("/operations/systems/settings")

	if av.Debug {
		av.log.Printf("POST %s\n%s\n", url, string(js))
	}

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	if err = av.parseResponse(resp, op); err != nil {
		return nil, err
	}

	return op, nil
}

// CreateSystem creates a new System on AirVantage. It returns a new System struct
// with updated information.
// Required fields in System: name, gateway
func (av *AirVantage) CreateSystem(system *System) (*System, error) {

	url := av.URL("systems")
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

	url := av.URL("systems/"+uid, "deleteGateway", deleteGateway, "deleteSubscription", deleteSubscription)

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

// ExportDataFromDevices downloads a DataAggregate of all the devices for a given company,
// in the given time interval. Optionally you can select the data points to return by
// setting `fields`, a comma-separated list of data IDs.
func (av *AirVantage) ExportDataFromDevices(companyUID, fields string, from, to time.Time) (DataAggregate, error) {

	resp, err := av.get("systems/data/fleet", "targetIds", companyUID, "dataIds", fields,
		"from", NewAVTime(from), "to", NewAVTime(to))
	if err != nil {
		return nil, err
	}

	data := DataAggregate{}
	if err = av.parseResponse(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// FindSystems is the generic method to find one or more systems.
// Parameters:
// - criteria is a map of field->value to filter the results
// - fields is a comma-separated list of fields to return (optional)
// - orderBy is a comma-separated list of fields to order the results (optional)
// You can limit the number of results (100 by default) by adding a criteria 'size'.
func (av *AirVantage) FindSystems(criteria url.Values, fields, orderBy string) ([]System, error) {
	if fields != "" {
		criteria.Set("fields", fields)
	}
	if orderBy != "" {
		criteria.Set("orderBy", orderBy)
	}

	resp, err := av.get("systems?" + criteria.Encode())
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
// - fields: a comma-separated list of fields to return (optional)
func (av *AirVantage) FindSystemByName(name, fields string) (*System, error) {
	criteria := url.Values{}
	criteria.Set("name", name)
	criteria.Set("size", "1")

	systems, err := av.FindSystems(criteria, fields, "")
	if err != nil || len(systems) == 0 {
		return nil, err
	}

	return &systems[0], err
}

// FindSystemByUID returns the System owning the given UID.
// Parameters:
// - fields: a comma-separated list of fields to return (optional)
func (av *AirVantage) FindSystemByUID(UID string) (*System, error) {

	resp, err := av.get("systems/" + UID)
	if err != nil {
		return nil, err
	}

	res := System{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type TsValue struct {
	Value     interface{} `json:"value"`
	Timestamp AVTime      `json:"timestamp"`
}

// GetLatestData returns the latests data points on a device, without querying it. You can
// optionally select which data to return by specifying a comma-separated list of data IDs.
func (av *AirVantage) GetLatestData(systemUID, dataIDs string) (map[string][]TsValue, error) {

	resp, err := av.get("systems/"+systemUID+"/data", "ids", dataIDs)
	if err != nil {
		return nil, err
	}

	res := map[string][]TsValue{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return res, nil
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
	url := av.URL("operations/systems/import")
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

	// Waiting for operation to finish
	if av.Debug {
		av.log.Println("waiting for systems import operation", res.Operation)
	}
	op, err := av.AwaitOperation(res.Operation, 1*time.Minute)
	if err != nil {
		return err
	}

	// Check if all the systems were created.
	if op.Counters.Failure > 0 {
		return fmt.Errorf("failed to create %d systems", op.Counters.Failure)
	}

	return nil
}

// InstallApplication installs or upgrades an application on a system
func (av *AirVantage) InstallApplication(appUID, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Application string `json:"application"`
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}
	body.Application = appUID

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/applications/install")

	if av.Debug {
		av.log.Printf("POST %s\n%s\n", url, string(js))
	}

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}
	av.log.Printf("POST %s\n%s\n", url, string(js))

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}
