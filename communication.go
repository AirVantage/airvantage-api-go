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

type SystemSecurityInfo struct {
	CommInfos []SystemCommInfo `json:"commInfo"`
}

type SystemCommInfo struct {
	SystemId          string `json:"systemId,omitempty"`
	CompanyId         string `json:"companyId,omitempty"`
	Lwm2mCommId       string `json:"lwm2mCommId,omitempty"`
	Lwm2mSecurityType string `json:"lwm2mSecurityType,omitempty"`
	Lwm2mPskIdentity  string `json:"lwm2mPskIdentity,omitempty"`
	Lwm2mPskSecretHex string `json:"lwm2mPskSecretHex,omitempty"`
}
