package airvantage

import (
	"encoding/json"
	"errors"
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

		if av.Debug {
			av.log.Printf("DBG operation %+v\n", op)
		}

		if op.State == "FINISHED" {
			return op, nil
		}

		if time.Now().Sub(start) > timeout {
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
