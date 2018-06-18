package airvantage

import (
	"os"
	"testing"
)

// The tests need to be sequential.
func TestMain(t *testing.T) {
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
		Gateway: Gateway{
			IMEI: "118218318418",
			Type: "api-gateway",
		},
	}
	sys, err := av.CreateSystem(&sysspec)
	if err != nil {
		t.Fatal(err)
	}

	defer av.DeleteSystem(sys.UID, true, false)

	if sys.Name != sysspec.Name {
		t.FailNow()
	}

}
