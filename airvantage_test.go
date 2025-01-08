package airvantage

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestMaskUrlParams(t *testing.T) {
	maskedParams := []string{"AUTHKEY"}

	rawUrl := "https://qa.airvantage.io/device/internal/securityinfo?id=123456789&type=IMEI&protocol=LWM2M&AUTHKEY=toto"
	expectedFilteredUrl := "https://qa.airvantage.io/device/internal/securityinfo?id=123456789&type=IMEI&protocol=LWM2M&AUTHKEY=***"

	filteredUrl := maskUrlParams(rawUrl, maskedParams)
	if filteredUrl != expectedFilteredUrl {
		t.Fatalf("expected: %v, got: %v", expectedFilteredUrl, filteredUrl)
	}

	t.Logf("filteredUrl: %v", filteredUrl)
}

func TestParseResponse(t *testing.T) {
	mockBody := `{
		"uid": "fffffdaaf4cf491cb865fd3e8263b164",
		"name": "Testing System",
		"state": "DEPLOYED",
		"gateway": {
			"state": "ACTIVE",
			"type": "",
			"serialNumber": "SN9876543210",
			"uid": "fffffe8af71b41778261f8b6299ea7d5",
			"serviceEndDate": null,
			"imei": "999993090184036",
			"macAddress": null,
			"gatewayModel": null
		},
		"communication": {
			"msci": null,
			"m3da": null,
			"rest": null,
			"mqtt": {
				"password": "toto"
			}
		},
		"applications": [
			{
				"type": "FX30S(WP7702)_LE",
				"category": "FIRMWARE",
				"uid": "ffff51c17bd3485d86f9a3d1f654481e",
				"revision": "21.05.0.54b96444",
				"name": "FX30S(WP7702)_R15.1.0.004"
			},
			{
				"type": "com.mqttclient.airvantage.app",
				"category": "APPLICATION",
				"uid": "ffff7da454524df0bb33faa50c9bd809",
				"revision": "1.0",
				"name": "Generic IMEI MQTT Client"
			}
		],
		"subscriptions": [
			{
				"state": "ACTIVE",
				"identifier": "89332401000017449238",
				"uid": "29119e46d94d4dba82d784913372002d",
				"eid": null,
				"networkIdentifier": "206018072254719",
				"mobileNumber": "337000023733800",
				"ipAddress": "100.71.238.72",
				"requestedIp": null,
				"operator": "SIERRA_WIRELESS",
				"productRefName": "SW Advanced2 2FF",
				"appletGeneration": "V4",
				"confType": "ADVANCED2",
				"formFactor": "2FF",
				"technology": "4G",
				"orderId": "SO-DAF13310",
				"activationImei": null,
				"imeiLock": null,
				"serviceEndDate": null,
				"serviceOfferId": "24c48c518a874d22a31e24195ff5c52f"
			}
		]
	}`
	mockUrl, _ := url.Parse("https://example.com/test?query=123")
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(mockBody))),
		Header:     make(http.Header),
		Request: &http.Request{
			Method: "GET",
			URL:    mockUrl,
		},
	}

	av := AirVantage{}
	sys := &System{}

	err := av.parseResponse(mockResponse, sys)
	if err != nil {
		t.Fatalf("error during parsing: %v", err)
	}

	if sys.UID != "fffffdaaf4cf491cb865fd3e8263b164" {
		t.Fatalf("invalid parsed system UID : expected: %v, got: %v", "fffffdaaf4cf491cb865fd3e8263b164", sys.UID)
	}
	if sys.Gateway.IMEI != "999993090184036" {
		t.Fatalf("invalid parsed IMEI : expected: %v, got: %v", "999993090184036", sys.Gateway.IMEI)
	}
	if len(sys.Applications) != 2 {
		t.Fatalf("invalid parsed applications number : expected: %v, got: %v", 2, len(sys.Applications))
	}
}

func TestParseResponseSerializedJava(t *testing.T) {
	mockBody := `[
    "com.sierrawireless.airvantage.services.communication.dto.SystemSecurityInfo",
    {
        "systemExist": true,
        "allowedToComm": true,
        "commInfo": [
            "com.sierrawireless.airvantage.services.communication.dto.SystemCommInfo",
            {
                "msciCommId": null,
                "msciPassword": null,
                "omadmCommId": null,
                "omadmSecurityType": "NONE",
                "omadmClientUsername": null,
                "omadmClientPassword": null,
                "omadmClientNonce": null,
                "omadmServerPassword": null,
                "omadmServerNonce": null,
                "awtda2CommId": null,
                "m3daCommId": null,
                "m3daSecurityType": "NONE",
                "m3daNonce": null,
                "m3daCipher": "NONE",
                "m3daSharedKey": null,
                "m3daCredential": null,
                "mqttCommId": null,
                "mqttPassword": null,
                "mqttSecurityType": "USER_PASSWORD",
                "restCommId": null,
                "restPassword": null,
                "systemId": "fffff989f4f64e01aa8f908386ba77f8",
                "companyId": "ffff8fc019c649bf8131afe3d3eb3663",
                "lwm2mCommId": "359146140001239",
                "lwm2mSecurityType": "PSK",
                "lwm2mPskIdentity": "FFFFE31330FD55964B866F02C1A0D6E7",
                "lwm2mPskSecretHex": "FFFF6206fbcdd0e02b91b55b2640b132",
                "lwm2mObserveSupported": false,
                "serialNumber": "EX4175253212BE",
                "imei": "359146140001239",
                "heartbeatSeconds": 630720000,
                "mqttBroker": false,
                "mqttCA": null
            }
        ]
    }]`
	mockUrl, _ := url.Parse("https://example.com/test?query=123")
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(mockBody))),
		Header:     make(http.Header),
		Request: &http.Request{
			Method: "GET",
			URL:    mockUrl,
		},
	}

	av := AirVantage{}
	res := []SystemSecurityInfo{}

	err := av.parseResponseSerializedJava(mockResponse, &res, javaObjectNamespaceSierra)
	if err != nil {
		t.Fatalf("error during parsing: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("invalid parsed System communication info number : expected: %v, got: %v", 1, len(res))
	}
	if len(res[0].CommInfos) != 1 {
		t.Fatalf("invalid parsed communication info number : expected: %v, got: %v", 1, len(res))
	}
	if res[0].CommInfos[0].Lwm2mPskIdentity != "FFFFE31330FD55964B866F02C1A0D6E7" {
		t.Fatalf("invalid LwM2M PSK ID : expected: %v, got: %v", "FFFFE31330FD55964B866F02C1A0D6E7", res[0].CommInfos[0].Lwm2mPskIdentity)
	}
}
