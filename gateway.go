package airvantage

// Gateway represents the network interface of a System.
type Gateway struct {
	UID          string   `json:"uid"`
	IMEI         string   `json:"imei,omitempty"`
	MacAddress   string   `json:",omitempty"`
	SerialNumber string   `json:",omitempty"`
	Type         string   `json:",omitempty"`
	Metadata     Metadata `json:",omitempty"`
	Labels       []string `json:",omitempty"`
	CreationDate AVTime   `json:",omitempty"`
	State        string   `json:",omitempty"`
}
