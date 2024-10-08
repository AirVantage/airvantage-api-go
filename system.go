package airvantage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

const (
	// regexp pattern to cleanup json from device/internal/securityinfo core API endpoint response
	javaObjectNamespaceSierra = `"com\.sierrawireless\.airvantage\.[^"]*",`
)

// A System descriptor.
type System struct {
	UID                 string            `json:"uid,omitempty"`
	Name                string            `json:"name,omitempty"`
	Type                string            `json:"type,omitempty"`
	State               string            `json:"state,omitempty"` // Deprecated
	LifeCycleState      string            `json:"lifeCycleState,omitempty"`
	ActivityState       string            `json:"activityState,omitempty"`
	CommStatus          string            `json:"comStatus,omitempty"`
	CreationDate        AVTime            `json:"creationDate,omitempty"`
	ActivationDate      AVTime            `json:"activationDate,omitempty"`
	LastStateChangeDate AVTime            `json:"lastStateChangeDate,omitempty"`
	LastCommDate        AVTime            `json:"lastCommDate,omitempty"`
	SyncStatus          string            `json:"syncStatus,omitempty"`
	LastSyncDate        AVTime            `json:"lastSyncDate,omitempty"`
	Labels              []string          `json:"labels,omitempty"`
	Gateway             *Gateway          `json:"gateway,omitempty"`
	Subscription        map[string]string `json:"subscription,omitempty"`
	Applications        []*Application    `json:"applications,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	Data                map[string]any    `json:"data,omitempty"`
	DataUsage           map[string]any    `json:"dataUsage,omitempty"`
	Offer               map[string]any    `json:"offer,omitempty"`
	Communication       *Communication    `json:"communication,omitempty"`
	Heatbeat            map[string]any    `json:"heartbeat,omitempty"`
	StatusReport        map[string]any    `json:"statusReport,omitempty"`
	Reports             []map[string]any  `json:"reports,omitempty"`
}

// A Datapoint retrieved from a System.
type Datapoint struct {
	ts AVTime
	v  any
}

type Info struct {
	Uid         string `json:"uid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Application string `json:"applicationId"`
}

type DataSet struct {
	Info          Info     `json:"info"`
	Configuration []string `json:"dataset"`
}

type AdvancedReports struct {
	Period  int  `json:"period"`
	DataSet Info `json:"dataset"`
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
	slog.Debug("HTTP POST", "url", url, "json", string(js))

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
	slog.Debug("HTTP POST", "url", url, "json", string(js))

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
	slog.Debug("HTTP POST", "url", url, "json", string(js))

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

// ActivateSystem activates a system
func (av *AirVantage) ActivateSystem(system *System) (string, error) {

	url := av.URL("operations/systems/activate")
	selection := struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
	}{}
	selection.Systems.UIDs = []string{system.UID}

	js, err := json.Marshal(selection)
	if err != nil {
		return "", err
	}
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// EditSystem updates the system
func (av *AirVantage) EditSystem(uid string, system *System) (*System, error) {

	url := av.URL("systems/" + uid)
	js, err := json.Marshal(system)
	if err != nil {
		return nil, err
	}
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	req, err := http.NewRequest("PUT", url, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := av.client.Do(req)
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
	slog.Debug("HTTP DELETE", "url", url)

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

	resp, err := av.getWithParams("systems", criteria)
	if err != nil {
		return nil, err
	}

	var page struct {
		Items []System `json:"items"`
	}
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

// GetSystemSecurityInfo returns the System communication info.
// Parameters:
// - authkey: authentication key for internal API
// - systemIdentifier: system identifier
// - secuType: communication identifier SERIAL_NUMBER, IMEI, MAC_ADDRESS, PSK_IDENTITY, CUSTOM
// - protocol: communication type MSCI, OMADM, AWTDA2, M3DA, REST, MQTT, LWM2M
func (av *AirVantage) GetSystemSecurityInfo(authkey string, systemIdentifier string, secuType string, protocol string) (*SystemSecurityInfo, error) {

	url := fmt.Sprintf("%s://%s/device/internal/securityinfo?id=%s&type=%s&protocol=%s&AUTHKEY=%s",
		av.baseURLv1.Scheme, av.baseURLv1.Host, systemIdentifier, secuType, protocol, authkey)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := av.client.Do(req)
	if err != nil {
		return nil, err
	}

	// get the raw response, which is java object serialization
	res := []SystemSecurityInfo{}
	if err = av.parseResponseSerializedJava(resp, &res, javaObjectNamespaceSierra); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("no %s commInfos found for system using %s %s", protocol, secuType, systemIdentifier)
	}

	return &res[0], nil
}

type TsValue struct {
	Value     any    `json:"value"`
	Timestamp AVTime `json:"timestamp"`
}

// GetLatestData V1 returns the latests data points on a device, without querying it. You can
// optionally select which data to return by specifying a comma-separated list of data IDs.
func (av *AirVantage) GetLatestData(systemUID, dataIDs string) (map[string][]TsValue, error) {
	var err error
	var resp *http.Response

	if dataIDs != "" {
		resp, err = av.get("systems/"+systemUID+"/data", "ids", dataIDs)
	} else {
		resp, err = av.get("systems/" + systemUID + "/data")
	}
	if err != nil {
		return nil, err
	}

	res := map[string][]TsValue{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type TsValueV2 struct {
	Value     any    `json:"v"`
	Timestamp AVTime `json:"ts"`
}

// GetLatestDataV2 returns the latests data points on a device, without querying it. You can
// optionally select which data to return by specifying a comma-separated list of data IDs.
func (av *AirVantage) GetLatestDataV2(systemUID, dataIDs string) (map[string][]TsValueV2, error) {
	var err error
	var resp *http.Response

	if dataIDs != "" {
		resp, err = av.getV2("systems/"+systemUID+"/data", "ids", dataIDs)
	} else {
		resp, err = av.getV2("systems/" + systemUID + "/data")
	}
	if err != nil {
		return nil, err
	}

	res := map[string][]TsValueV2{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type UnityConf struct {
	Current struct {
		Value     any    `json:"value"`
		Timestamp AVTime `json:"ts"`
	} `json:"current"`
	Action struct {
		OperationID string `json:"operationId"`
		TaskID      string `json:"taskId"`
		Value       any    `json:"value"`
		ValueType   string `json:"valueType"`
		Timestamp   AVTime `json:"ts"`
		Status      string `json:"status"`
	} `json:"action"`
}

// GetUnityConfig returns the configuration of a Unity gateways (last datapoints, pending actions...)
func (av *AirVantage) GetUnityConfig(systemUID string) (map[string]UnityConf, error) {

	resp, err := av.get("unity/" + systemUID + "/conf")
	if err != nil {
		return nil, err
	}

	res := map[string]UnityConf{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type UnityCommand struct {
	OperationID string `json:"operationId"`
	TaskID      string `json:"taskId"`
	Timestamp   AVTime `json:"ts"`
	Status      string `json:"status"`
}

// GetUnityCommand returns the command status of a Unity gateways
func (av *AirVantage) GetUnityCommand(systemUID string) (map[string]UnityCommand, error) {

	resp, err := av.get("unity/" + systemUID + "/command")
	if err != nil {
		return nil, err
	}

	res := map[string]UnityCommand{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// DismissUnityCommand returns the command status of a Unity gateways
func (av *AirVantage) DismissUnityCommand(systemUID string, commandID string) error {

	body := struct {
		CommandIDS []string `json:"commandIds"`
	}{CommandIDS: []string{commandID}}

	js, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	url := av.URL("unity/" + systemUID + "/command/dismisserror")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	_, err = av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return err
	}

	return nil
}

// ImportSystemsDefaults provides optional information to the
// ImportSystems operation.
type ImportSystemsDefaults struct {
	// Send an email notification when the operation finishes.
	Notify bool `json:"notify,omitempty"`
	// Callback URL when the operation finishes.
	Callback string `json:"callback,omitempty"`
	// Default application IDs linked to the systems.
	DefaultApplicationsUID []string `json:"defaultApplications,omitempty"`
	// Default system state.
	DefaultState string `json:"defaultState,omitempty"`
	// Default system type.
	DefaultType string `json:"defaultType,omitempty"`
}

// ImportSystems creates a batch of systems using data provided in CSV format.
// To reduce repetitions in the CSV, you should provide `defaults` that will
// be applied for each system. The default timeout is 5 minutes.
func (av *AirVantage) ImportSystems(csv io.Reader, defaults *ImportSystemsDefaults, timeout time.Duration) error {

	if csv == nil {
		return fmt.Errorf("csv reader is nil")
	}
	if defaults == nil {
		defaults = &ImportSystemsDefaults{}
	}
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Create a multi-part request.
	var bb bytes.Buffer
	multi := multipart.NewWriter(&bb)

	// CSV part
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="csv"; filename="file.csv"`)
	header.Set("Content-Type", "text/csv")
	partWriter, _ := multi.CreatePart(header)
	if _, err := io.Copy(partWriter, csv); err != nil {
		return fmt.Errorf("ImportSystems: %s", err)
	}

	// JSON part
	header = make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="parameters"; filename="parameters.json"`)
	header.Set("Content-Type", "application/json")
	partWriter, _ = multi.CreatePart(header)
	js, err := json.Marshal(defaults)
	if err != nil {
		return fmt.Errorf("ImportSystems: %s", err)
	}
	partWriter.Write(js)

	multi.Close()

	req, err := http.NewRequest("POST", av.URL("operations/systems/import"), &bb)
	if err != nil {
		return fmt.Errorf("ImportSystems: %s", err)
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
	slog.Debug("waiting for systems import operation", "uid", res.Operation)

	op, err := av.AwaitOperation(res.Operation, timeout)
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
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// RetrieveData launch an operation to read the given paths on the system
func (av *AirVantage) RetrieveData(paths []string, protocol string, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Data     []string `json:"data"`
		Protocol string   `json:"protocol"`
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}
	body.Data = paths
	if protocol != "" {
		body.Protocol = protocol
	}

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/data/retrieve")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// Configure Communication launch an operation to configure the communication on the system.
func (av *AirVantage) ConfigureCommunication(hbState string, hbPeriod int, srState string, srPeriod int, systemsUID []string, reports []AdvancedReports) (string, error) {

	type HeartBeat struct {
		State      string `json:"state"`
		Period     int    `json:"period"`
		ServerOnly bool   `json:"serverOnly"`
	}

	type StatusReport struct {
		State  string `json:"state"`
		Period int    `json:"period"`
	}

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		HeartBeat       HeartBeat         `json:"heartbeat"`
		StatusReport    StatusReport      `json:"statusReport"`
		AdvancedReports []AdvancedReports `json:"reports"`
	}

	var body jsonBody
	if len(systemsUID) > 0 {
		body.Systems.UIDs = systemsUID
		if hbPeriod != 0 {
			body.HeartBeat.State = hbState
			body.HeartBeat.Period = hbPeriod
		}

		if srPeriod != 0 {
			body.StatusReport.State = srState
			body.StatusReport.Period = srPeriod
		}

		if len(reports) > 0 {
			body.AdvancedReports = reports
		}
	}
	body.HeartBeat.ServerOnly = false

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	ccUrl := av.URL("operations/systems/configure")
	slog.Debug("HTTP POST", "url", ccUrl, "json", string(js))

	resp, err := av.client.Post(ccUrl, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

func (av *AirVantage) CreateDataset(name string, description string, configuration []string, appId string) (*DataSet, error) {

	var dataset DataSet
	dataset.Info.Name = name
	dataset.Info.Description = description
	dataset.Configuration = configuration
	dataset.Info.Application = appId

	js, err := json.Marshal(&dataset)
	if err != nil {
		return nil, err
	}

	ccUrl := av.URL("/api/v2/datasets")
	slog.Debug("HTTP POST", "url", ccUrl, "json", string(js))

	resp, err := av.client.Post(ccUrl, "application/json", bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	slog.Info("Create dataset response", "resp", resp)

	res := &DataSet{}
	if err = av.parseResponse(resp, &res); err != nil {
		return nil, err
	}
	slog.Info("Create dataset response parsed", "res", res)

	return res, nil
}

// ApplySettings launch an operation to write/delete the given settings on the system
func (av *AirVantage) ApplySettings(settings map[string]any, delete []string, protocol, systemUID string) (string, error) {

	type Setting struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Settings []Setting `json:"settings"`
		Delete   []string  `json:"deleteSettings"`
		Protocol string    `json:"protocol"`
		Reboot   bool      `json:"reboot"`
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}
	body.Settings = make([]Setting, len(settings))
	if len(delete) > 0 {
		body.Delete = delete
	}
	var i = 0
	for k, v := range settings {
		body.Settings[i] = Setting{Key: k, Value: v}
		i++
	}
	if protocol != "" {
		body.Protocol = protocol
	}
	body.Reboot = false

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/settings")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// SendCommand launch an operation to run the given command and parameters on the system
func (av *AirVantage) SendCommand(commandID string, parameters map[string]any, protocol, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		CommandID  string         `json:"commandId"`
		Parameters map[string]any `json:"parameters"`
		Protocol   string         `json:"protocol"`
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}
	body.CommandID = commandID
	body.Parameters = parameters

	if protocol != "" {
		body.Protocol = protocol
	}

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/command")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// SendFile launches an operation to send the given file to a system
func (av *AirVantage) SendFile(fileID, target, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		FileID string `json:"file"`
		Target string `json:"target"`
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}
	body.FileID = fileID
	body.Target = target

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/file/send")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// Reboot launch an operation to run a reboot on the given system
func (av *AirVantage) Reboot(action string, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Action *string `json:"action"` //optional, null by default
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}

	if action != "" {
		body.Action = &action
	}

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/reboot")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}

// Reset launch an operation to run a factory Reset on the given system
func (av *AirVantage) Reset(action string, systemUID string) (string, error) {

	type jsonBody struct {
		Systems struct {
			UIDs []string `json:"uids"`
		} `json:"systems"`
		Action *string `json:"action"` //optional, null by default
	}
	var body jsonBody
	body.Systems.UIDs = []string{systemUID}

	if action != "" {
		body.Action = &action
	}

	js, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	url := av.URL("operations/systems/reset")
	slog.Debug("HTTP POST", "url", url, "json", string(js))

	resp, err := av.client.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		return "", err
	}

	res := struct{ Operation string }{}
	if err = av.parseResponse(resp, &res); err != nil {
		return "", err
	}
	return string(res.Operation), nil
}
