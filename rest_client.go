package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type restClient struct {
	baseUrl string
}

func newRestClient(baseUrl string) *restClient {
	return &restClient{baseUrl: baseUrl}
}

func (client *restClient) doGet(path string, responseTarget interface{}) error {
	res, err := http.Get(client.baseUrl + path)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return errors.New("http response failed with " + strconv.Itoa(res.StatusCode))
	}
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusCreated {
		return nil
	}
	if responseTarget == nil {
		return nil
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, responseTarget)
	if err != nil {
		return err
	}
	return nil
}

func (client *restClient) doPost(path string, requestBody interface{}, responseTarget interface{}) error {
	requestData := make([]byte, 0)
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		requestData = data
	}
	res, err := http.Post(client.baseUrl+path, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return errors.New("http response failed with " + strconv.Itoa(res.StatusCode))
	}
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusCreated {
		return nil
	}
	if responseTarget == nil {
		return nil
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, responseTarget)
	if err != nil {
		return err
	}
	return nil
}

func (client *restClient) doDelete(path string) error {
	req, err := http.NewRequest("DELETE", client.baseUrl+path, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return errors.New("http response failed with " + strconv.Itoa(res.StatusCode))
	}
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusCreated {
		return nil
	}
	return nil
}
