package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AirVantage/airvantage-api-go"
)

var (
	avURL        = flag.String("av", "https://qa.airvantage.io", "AV backend base URL")
	clientID     = flag.String("apikey", "aa5551d0d2eb4fd0a3f4e09129114ac9", "API client ID")
	clientSecret = flag.String("apisecret", "af4f646d45d94d6a979e10ff50f7f068", "API client secret")
	login        = flag.String("login", "", "AirVantage login email")
	password     = flag.String("passwd", "", "AirVantage password")
	companyID    = flag.String("company", "8132c4cdccf34776ba0aa8e361de1952", "Company UID to use (Capsule Corp by default)")
	appID        = flag.String("appid", "6a886cb764cc44ccb105fefd36ea6828", "Application ID linked to the devices")
	deviceName   = flag.String("devname", "avfleet", "Prefix used for the devices name, serial and imei")
	quantity     = flag.Int("qty", 10, "Quantity of devices to create")
)

func genCSV() *bytes.Buffer {
	bb := new(bytes.Buffer)
	writer := csv.NewWriter(bb)

	writer.Write([]string{"NAME", "LABELS", "GATEWAY[IMEI]", "GATEWAY[SERIAL NUMBER]"})

	for i := 1; i <= *quantity; i++ {
		id := fmt.Sprintf("%s%08d", *deviceName, i)
		writer.Write([]string{id, *deviceName, id, id})
	}

	writer.Flush()

	return bb
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *login == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	av, err := airvantage.NewClient(*avURL, *clientID, *clientSecret, *login, *password)
	if err != nil {
		log.Fatal(err)
	}
	av.CompanyUID = *companyID

	defaults := &airvantage.ImportSystemsDefaults{
		DefaultApplicationsUID: []string{*appID},
		DefaultState:           "READY",
		DefaultType:            *deviceName,
	}

	if err = av.ImportSystems(genCSV(), defaults, 0); err != nil {
		log.Fatal(err)
	}

	log.Printf("Created %d devices successfully.\n", *quantity)
}
