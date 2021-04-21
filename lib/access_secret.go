package lib

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type AccessSecretInput struct {
	PeerNodeName string `json:"peerNodeName"`
}

type AccessSecretOutput struct {
	Secret string `json:"secret"`
}

// AccessSecret return an access secret for a service provider
func AccessSecret(peerNodeName string, controller *Controller) (*AccessSecretOutput, error) {
	requestUrl := GetEndpointUrl(AccessSecretEndpoint)
	input := AccessSecretInput{PeerNodeName: peerNodeName}
	resp, err := controller.SendRequest(requestUrl, &input, http.MethodPost, true)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s  %s", resp.StatusCode, resp.Status, "not authorized"))
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		respBodyString := string(respBody)
		if respBodyString != "" {
			return nil, errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s, Body: %s", resp.StatusCode, resp.Status, respBodyString))
		}
		return nil, errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s", resp.StatusCode, resp.Status))
	}
	var accessSecretOutput AccessSecretOutput
	respBody, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &accessSecretOutput); err != nil {
		return nil, err
	}
	return &accessSecretOutput, nil
}
