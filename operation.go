package airvantage

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrOperationExpired is returned when an operations timeouts.
var ErrOperationExpired = errors.New("operation expired")

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

// NewOperation creates a new, empty Operation with the given ID.
func NewOperation(uid string) *Operation {
	return &Operation{UID: uid}
}

// AwaitOperation blocks until the operation is finished or expired.
// When it returns, the operation's details are updated.
func (av *AirVantage) AwaitOperation(op *Operation) error {
	for {
		newOp, err := av.GetOperation(op.UID)
		if err != nil {
			return err
		}

		if av.Debug {
			av.log.Printf("DBG operation %+v\n", newOp)
		}

		if newOp.State == "FINISHED" {
			*op = *newOp
			break
		}

		if time.Now().After(newOp.Timeout.Time()) {
			*op = *newOp
			return ErrOperationExpired
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}

// GetOperation retrieves details about an Operation.
func (av *AirVantage) GetOperation(uid string) (*Operation, error) {
	resp, err := av.get("/operations/" + uid)
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	if err = av.parseResponse(resp, op); err != nil {
		return nil, err
	}

	return op, nil
}
