package nights_watch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type EthRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type EthResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *EthError       `json:"error"`
}

func (err EthError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

// EthError - ethereum error
type EthError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Call returns raw response of method call
func call(client *http.Client, url, method string, target interface{}) error {
	request := EthRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	response, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	resp := new(EthResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return *resp.Error
	}

	return json.Unmarshal(resp.Result, target)
}

func ParseInt(value string) (int, error) {
	i, err := strconv.ParseInt(strings.TrimPrefix(value, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}
