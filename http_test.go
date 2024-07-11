package sdk

import (
	"encoding/json"
	"github.com/runetid/go-sdk/models"
	"io"
	"net"
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

func TestNetworks(t *testing.T) {

	_, local, _ := net.ParseCIDR("172.31.0.8/16")
	_, remote, _ := net.ParseCIDR("172.31.0.7/16")

	if intersect(local, remote) == false {
		t.Fatal("Wrong addresses intersect")
	}
}

func TestRawFetchModel(t *testing.T) {
	model := models.User{}

	url := "http://localhost:57449/internal/byToken/peOWZrIEzHmbQYjNtwWb"

	model, err := RawFetchModel[models.User]("GET", url, nil, "traceId", model)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if model.Id != 66427 || model.RunetId != 87610 {
		t.Errorf("Expected model with id 66427, got %v", model)
	}
}
