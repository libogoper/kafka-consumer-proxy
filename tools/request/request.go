package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func newError(msg string, v ...interface{}) error {
	if v == nil {
		return errors.New(msg)
	}
	return errors.New(fmt.Sprintf(msg, v))
}
func Do(method, url string, headers map[string]string, sendData []byte) ([]byte, error) {
	timeout := time.Second * 5
	var body io.Reader
	//send Request
	if sendData != nil {
		body = bytes.NewBuffer(sendData)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, newError("http.NewRequest failed with %s", err.Error())
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)
	if err != nil {
		return nil, newError("http.client.Do() failed with %s", err.Error())
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newError("io.ReadAll() failed with %s", err.Error())
	}
	//check response params
	if resp.StatusCode != 200 || content == nil {
		return nil, newError("curl_fail with status %s", resp.Status)
	}
	return content, nil
}
