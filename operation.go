package airvantage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/url"
	"strings"
	"time"
)

// ErrWaitFinishedOperationTimeout is returned when waiting too long for a finished operation.
var ErrWaitFinishedOperationTimeout = errors.New("wait.operation.timeout")

type OperationCounters struct {
	BeingCancelled int
	Cancelled      int
	Failure        int
	InProgress     int
	Pending        int
	Success        int
}

type jsonCounter struct {
	State string
	Count int
}

func (oc *OperationCounters) UnmarshalJSON(b []byte) error {
	var counters []jsonCounter
	if err := json.Unmarshal(b, &counters); err != nil {
		return err
	}

	for _, cnt := range counters {
		switch cnt.State {
		case "BEING_CANCELLED":
			oc.BeingCancelled = cnt.Count
		case "CANCELLED":
			oc.Cancelled = cnt.Count
		case "FAILURE":
			oc.Failure = cnt.Count
		case "IN_PROGRESS":
			oc.InProgress = cnt.Count
		case "PENDING":
			oc.Pending = cnt.Count
		case "SUCCESS":
			oc.Success = cnt.Count
		}
	}

	return nil
}

func (oc OperationCounters) MarshalJSON() ([]byte, error) {
	var counters = []jsonCounter{
		{"BEING_CANCELLED", oc.BeingCancelled},
		{"CANCELLED", oc.Cancelled},
		{"FAILURE", oc.Failure},
		{"IN_PROGRESS", oc.InProgress},
		{"PENDING", oc.Pending},
		{"SUCCESS", oc.Success},
	}
	return json.Marshal(counters)
}

// An Operation descriptor.
type Operation struct {
	UID      string `json:"uid"`
	State    string
	Timeout  AVTime `json:"timeoutDate,omitempty"`
	Counters OperationCounters
}

// AwaitOperation blocks until the operation is finished or expired.
func (av *AirVantage) AwaitOperation(opUID string, timeout time.Duration) (*Operation, error) {
	start := time.Now()
	for {
		op, err := av.GetOperation(opUID)
		if err != nil {
			return nil, err
		}

		slog.Debug("Waiting Operation", "uid", op)

		if op.State == "FINISHED" {
			return op, nil
		}

		if time.Since(start) > timeout {
			return op, ErrWaitFinishedOperationTimeout
		}

		time.Sleep(5 * time.Second)
	}
}

// GetOperation retrieves details about an Operation.
func (av *AirVantage) GetOperation(uid string) (*Operation, error) {
	resp, err := av.get("operations/" + uid)
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	if err = av.parseResponse(resp, op); err != nil {
		return nil, err
	}

	return op, nil
}

// FindOperations is the generic method to find one or more systems.
// Parameters:
// - criteria is a map of field->value to filter the results
// - fields is a comma-separated list of fields to return (optional)
// - orderBy is a comma-separated list of fields to order the results (optional)
// You can limit the number of results (100 by default) by adding a criteria 'size'.
func (av *AirVantage) FindOperations(criteria url.Values, fields, orderBy string) ([]Operation, error) {
	if fields != "" {
		criteria.Set("fields", fields)
	}
	if orderBy != "" {
		criteria.Set("orderBy", orderBy)
	}

	resp, err := av.getWithParams("operations", criteria)
	if err != nil {
		return nil, err
	}

	var page struct {
		Items []Operation `json:"items"`
	}
	if err = av.parseResponse(resp, &page); err != nil {
		return nil, err
	}

	return page.Items, nil
}

// CancelOperation cancels the operation with the given UID.
func (av *AirVantage) CancelOperation(opUID string) (*Operation, error) {

	resp, err := av.client.Post(av.URL(fmt.Sprintf("operations/%s/cancel", opUID)), "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	if err = av.parseResponse(resp, &op); err != nil {
		return nil, err
	}
	return op, nil
}

// GetOperationUnsignedPayload retrieves the operation unsigned payload as a JSON string
func (av *AirVantage) GetOperationUnsignedPayload(uid string) (string, error) {
	resp, err := av.get(fmt.Sprintf("operations/%s/unsignedpayload", uid))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid response when retrieveing unsignedpayload: %+v", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ApproveOperation approves the operation with the given UID.
func (av *AirVantage) ApproveOperation(opUID, base64Signature, algorithm string, certificateChain io.Reader) (err error) {

	values := make(map[string]io.Reader)
	values["cert"] = certificateChain
	values["signature"] = strings.NewReader(base64Signature)
	values["algorithm"] = strings.NewReader(algorithm)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if key == "cert" {
			// Add certificate
			if fw, err = w.CreateFormFile(key, "cert-chain.pem"); err != nil {
				return err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}
	}
	w.Close()

	resp, err := av.client.Post(av.URL(fmt.Sprintf("operations/%s/approve", opUID)), w.FormDataContentType(), &b)
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response for approve API call: %+v", resp)
	}
	return err
}
