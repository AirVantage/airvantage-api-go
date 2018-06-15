package airvantage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	js := fmt.Sprintf(`{"date":%d}`, time.Now().UnixNano()/1000)

	res := struct{ Date AVTime }{}
	if err := json.Unmarshal([]byte(js), &res); err != nil {
		t.Error(err)
		return
	}

	if res.Date.Time().Unix() != int64(res.Date)/1000 {
		t.Fail()
	}
}

func TestEncode(t *testing.T) {
	avt := AVTime(time.Now().UnixNano() / 1000)
	js, err := json.Marshal(avt)
	if err != nil {
		t.Error(err)
		return
	}

	i, err := strconv.ParseInt(string(js), 10, 64)
	if err != nil {
		t.Error(err)
		return
	}
	if i != int64(avt) {
		t.Fail()
	}
}
