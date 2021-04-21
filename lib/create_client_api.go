package lib

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type CreateClient struct {
	NodeName string `json:"nodeName"`
}

type ISEClient struct {
	NodeName string `json:"nodeName"`
	Password string `json:"password"`
	UserName string `json:"userName"`
}

type AccountActivate struct {
	AccountState string `json:"accountState"`
	Version      string `json:"version"`
}

// Create create a ISE Client Account
func (c *CreateClient) Create(controller *Controller) (*ISEClient, error) {
	requestUrl := GetEndpointUrl(PxGridCreateClientEndPoint)
	resp, err := controller.SendRequest(requestUrl, c, http.MethodPost, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
	var iseClient ISEClient
	if err := json.NewDecoder(resp.Body).Decode(&iseClient); err != nil {
		return nil, err
	}
	return &iseClient, nil
}

// AccountActivate Activate ISE Client Account
func (c *CreateClient) AccountActivate(controller *Controller) (*AccountActivate, error) {
	requestUrl := GetEndpointUrl(PxGridAccountActivateEndPoint)
	resp, err := controller.SendRequest(requestUrl, c, http.MethodPost, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
	var accountActivate AccountActivate
	if err := json.NewDecoder(resp.Body).Decode(&accountActivate); err != nil {
		return nil, err
	}
	return &accountActivate, nil

}
