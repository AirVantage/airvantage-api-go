package airvantage

// Gateway represents the network interface of a System.
type Gateway struct {
	UID          string   `json:"uid,omitempty"`
	IMEI         string   `json:"imei,omitempty"`
	MacAddress   string   `json:"macAddress,omitempty"`
	SerialNumber string   `json:"serialNumber,omitempty"`
	Type         string   `json:"type,omitempty"`
	Metadata     Metadata `json:"metadata,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	CreationDate AVTime   `json:"creationDate,omitempty"`
	State        string   `json:"state,omitempty"`
}
