package sdk

import (
	"encoding/json"
	"github.com/runetid/go-sdk/models"
	"io"
	"net/http"
	"testing"
)

func TestApiAccount(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://tmpapi.runet.id/apiaccount/check", nil)

	q := req.URL.Query()
	q.Add("ApiKey", "")
	q.Add("Hash", "")
	q.Add("Time", "")
	q.Add("Origin", "https://tmpapi.runet.id")
	req.URL.RawQuery = q.Encode()

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		t.Log(err)
	}

	var response models.ApiAccountResponse
	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		t.Log("Can not unmarshal api account response " + string(body))

	}
}
