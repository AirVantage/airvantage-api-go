package airvantage

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	TestingHost    string = "qa.airvantage.io"
	companyUID     string = "8f70416f52c04483a74e4baf12496f0e"
)

type AvCredentials struct {
	apiKey    string
	apiSecret string
}

func getTestingCreds() (AvCredentials, error) {
	creds := AvCredentials{
		apiKey:    os.Getenv("API_KEY"),
		apiSecret: os.Getenv("API_SECRET"),
	}

	if len(creds.apiKey) == 0 || len(creds.apiSecret) == 0 {
		return creds, fmt.Errorf("missing credentials for integration tests")
	}

	return creds, nil
}

// The tests need to be sequential.

func TestSystem(t *testing.T) {
    const (
    	// system parameters
    	system      string = "api test"
    	gatewayIMEI string = "118218318418"
    	gatewayType string = "api-gateway"
    )

	creds, err := getTestingCreds()
	if err != nil {
		t.Fatal(err)
	}

	av, err := NewClient(
		TestingHost,
		creds.apiKey,
		creds.apiSecret,
	)
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true

	// Create a new System
	sysSpec := System{
		Name: system,
		Gateway: &Gateway{
			IMEI: gatewayIMEI,
			Type: gatewayType,
		},
	}
	sys, err := av.CreateSystem(&sysSpec)
	if err != nil {
		t.Fatal(err)
	}
	defer av.DeleteSystem(sys.UID, true, false)

	sys, err = av.FindSystemByUID(sys.UID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Found: %+v", sys)

	if sys.Name != sysSpec.Name {
		t.FailNow()
	}
}

func TestFindAppByTypeRev(t *testing.T) {
     const (
        // application parameters
        appType     string = "test-mqtt"
        appRev      string = "0.1"
     )

	creds, err := getTestingCreds()
	if err != nil {
		t.Fatal(err)
	}

	av, err := NewClient(
		TestingHost,
		creds.apiKey,
		creds.apiSecret,
	)
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true

	app, err := av.FindAppByTypeRev(appType, appRev)
	if err != nil {
		t.Fatal(err)
	}

	if app == nil {
		t.Fatal("app not found")
	}

	t.Logf("Found: %+v", app)
}

func TestInstallApp(t *testing.T) {
    t.SkipNow()

    const (
        // application parameters
        appUID      string = "c634dd2234714578ad286c04e038f5b2"

        // system parameters
        systemUID   string = "42be9f3a82d94da5bc3d44af67138092"
    )

	creds, err := getTestingCreds()
	if err != nil {
		t.Fatal(err)
	}

	av, err := NewClient(
		TestingHost,
		creds.apiKey,
		creds.apiSecret,
	)
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	av.CompanyUID = companyUID

	opUID, err := av.InstallApplication(appUID, systemUID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Install op UID: %+v", opUID)

	op, err := av.AwaitOperation(opUID, 10*time.Second)
	if err != ErrWaitFinishedOperationTimeout {
		// timeout because the operation cannot be finished
		t.Fatalf("expected: %v, got: %v", ErrWaitFinishedOperationTimeout, err)
	}
	t.Logf("Install op: %+v", op)
}

func TestGetOperation(t *testing.T) {
	t.SkipNow()

    const (
        opUID       string = "46be2ae142dc4fd993819a38b2937e2d"
    )

	creds, err := getTestingCreds()
	if err != nil {
		t.Fatal(err)
	}

	av, err := NewClient(
		TestingHost,
		creds.apiKey,
		creds.apiSecret,
	)
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	//av.CompanyUID = companyUID

	op, err := av.GetOperation(opUID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Op: %+v", op)
}

func TestGetLatestData(t *testing.T) {
	t.SkipNow()

	const (
        systemUID      string = "5bdacf411b5d4603a6d13f099a9ca5ba"
        dataIDs        string = "DM.SW.VER"
    )

	creds, err := getTestingCreds()
	if err != nil {
		t.Fatal(err)
	}

	av, err := NewClient(
		TestingHost,
		creds.apiKey,
		creds.apiSecret,
	)
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	//av.CompanyUID = companyUID

	res, err := av.GetLatestData(systemUID, dataIDs)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("res: %+v", res)
}
