package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"terraform-provider-opensearch/signing"
)

type Client struct {
	settingsEndpoint string
	httpClient       *http.Client
	requestSigner    signing.Signer
}

func NewClient(endpoint string, requestSigner signing.Signer) *Client {
	return &Client{
		settingsEndpoint: endpoint + "/_cluster/settings",
		httpClient:       http.DefaultClient,
		requestSigner:    requestSigner,
	}
}

type settingsMessage struct {
	Persistent *PersistentSettings `json:"persistent"`
}

type PersistentSettings struct {
	AutoCreateIndex *string `json:"action.auto_create_index"`
}

func (client *Client) Fetch() (*PersistentSettings, error) {
	req, err := http.NewRequest(http.MethodGet, client.settingsEndpoint+"?flat_settings", nil)
	if err != nil {
		return nil, err
	}

	err = client.requestSigner.Sign(req, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		responseBytes, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf(
				"unable to fetch cluster settings: %d(%s) - %s",
				res.StatusCode,
				res.Status,
				string(responseBytes))
		} else {
			return nil, fmt.Errorf(
				"unable to fetch cluster settings: %d(%s)",
				res.StatusCode,
				res.Status)
		}
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	message := settingsMessage{}
	err = json.Unmarshal(resBytes, &message)
	if err != nil {
		return nil, err
	} else if message.Persistent == nil {
		return nil, fmt.Errorf("expected object with persistent setting, got nil")
	}

	return message.Persistent, nil
}

func (client *Client) Update(settings *PersistentSettings) error {
	bodyBytes, err := json.Marshal(settingsMessage{Persistent: settings})
	if err != nil {
		return err
	}

	body := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(http.MethodPut, client.settingsEndpoint, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	err = client.requestSigner.Sign(req, body)
	if err != nil {
		return err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		responseBytes, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return fmt.Errorf(
				"unable to update cluster settings: %d(%s) - %s",
				res.StatusCode,
				res.Status,
				string(responseBytes))
		} else {
			return fmt.Errorf(
				"unable to update cluster settings: %d(%s)",
				res.StatusCode,
				res.Status)
		}
	}

	return nil
}
