package lib

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type ServiceLookupInput struct {
	Name string `json:"name"`
}

type ServiceLookupOutput struct {
	Services []Services `json:"services"`
}

type Services struct {
	Name       string            `json:"name,omitempty"`
	NodeName   string            `json:"nodeName,omitempty"`
	Properties ServiceProperties `json:"properties,omitempty"`
}

type ServiceProperties struct {
	SessionTopic    string `json:"sessionTopic,omitempty"`
	GroupTopic      string `json:"groupTopic,omitempty"`
	WsPubSubService string `json:"wsPubsubService,omitempty"`
	RestBaseURL     string `json:"restBaseURL,omitempty"`
	RestBaseUrl     string `json:"restBaseUrl,omitempty"`
	WsUrl           string `json:"wsUrl"`
}

func ServiceLookupRequest(serviceName string, controller *Controller) (*ServiceLookupOutput, error) {
	input := ServiceLookupInput{Name: serviceName}
	requestUrl := GetEndpointUrl(ServiceLookup)
	resp, err := controller.SendRequest(requestUrl, &input, http.MethodPost, true)
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
	var serviceLookupOutput ServiceLookupOutput
	respBody, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &serviceLookupOutput); err != nil {
		return nil, err
	}
	return &serviceLookupOutput, nil
}
