package airvantage

type ComProto struct {
	Host                 string `json:"host,omitempty"`
	User                 string `json:"user,omitempty"`
	Password             string `json:"password,omitempty"`
	RegistrationPassword string `json:"registrationPassword,omitempty"`
}

type Communication struct {
	MSCI ComProto `json:"msci,omitempty"`
	M3DA ComProto `json:"m3da,omitempty"`
	REST ComProto `json:"rest,omitempty"`
	MQTT ComProto `json:"mqtt,omitempty"`
}
