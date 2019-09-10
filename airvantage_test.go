package airvantage

import (
	"os"
	"testing"
	"time"
)

// The tests need to be sequential.
func TestSystem(t *testing.T) {
	av, err := NewClient("qa.airvantage.io",
		os.Getenv("API_KEY"), os.Getenv("API_SECRET"),
		os.Getenv("AV_LOGIN"), os.Getenv("AV_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true

	// Create a new System
	sysspec := System{
		Name: "api test",
		Gateway: &Gateway{
			IMEI: "118218318418",
			Type: "api-gateway",
		},
	}
	sys, err := av.CreateSystem(&sysspec)
	if err != nil {
		t.Fatal(err)
	}
	defer av.DeleteSystem(sys.UID, true, false)

	sys, err = av.FindSystemByUID(sys.UID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Found: %+v", sys)

	if sys.Name != sysspec.Name {
		t.FailNow()
	}
}

func TestFindAppByTypeRev(t *testing.T) {

	av, err := NewClient("https://qa.airvantage.io",
		os.Getenv("API_KEY"), os.Getenv("API_SECRET"),
		os.Getenv("AV_LOGIN"), os.Getenv("AV_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true

	app, err := av.FindAppByTypeRev("test-mqtt", "0.1")
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

	av, err := NewClient("https://qa.airvantage.io",
		os.Getenv("API_KEY"), os.Getenv("API_SECRET"),
		os.Getenv("AV_LOGIN"), os.Getenv("AV_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	av.CompanyUID = "8f70416f52c04483a74e4baf12496f0e"

	opUID, err := av.InstallApplication("c634dd2234714578ad286c04e038f5b2", "42be9f3a82d94da5bc3d44af67138092")
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

	av, err := NewClient("https://qa.airvantage.io",
		os.Getenv("API_KEY"), os.Getenv("API_SECRET"),
		os.Getenv("AV_LOGIN"), os.Getenv("AV_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	//av.CompanyUID = "8f70416f52c04483a74e4baf12496f0e"

	op, err := av.GetOperation("46be2ae142dc4fd993819a38b2937e2d")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Op: %+v", op)
}

func TestGetLatestData(t *testing.T) {

	t.SkipNow()

	av, err := NewClient("https://qa.airvantage.io",
		os.Getenv("API_KEY"), os.Getenv("API_SECRET"),
		os.Getenv("AV_LOGIN"), os.Getenv("AV_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}

	av.Debug = true
	//av.CompanyUID = "8f70416f52c04483a74e4baf12496f0e"

	res, err := av.GetLatestData("5bdacf411b5d4603a6d13f099a9ca5ba", "DM.SW.VER")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("res: %+v", res)
}
