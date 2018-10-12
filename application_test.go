package airvantage

import (
	"os"
	"testing"
)

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
