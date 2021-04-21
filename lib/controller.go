package lib

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type Controller struct {
	config    *Config
	client    *http.Client
	tlsConfig *tls.Config
}

// GetTlsConfig return the controller TLS config
func (c *Controller) GetTlsConfig() *tls.Config {
	return c.tlsConfig
}

// NewControl create a new controller for ISE API
func NewControl(config *Config) (control *Controller, err error) {
	tlsConfig, err := config.GetTLSConfig()
	if err != nil {
		return
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	control = &Controller{
		config:    config,
		client:    &http.Client{Transport: transport},
		tlsConfig: tlsConfig,
	}
	return
}

// SendRequest Send request to ISE API
func (c *Controller) SendRequest(url string, requestBody interface{}, requestMethod string, requireAuth bool) (*http.Response, error) {
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeoutValue*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, requestMethod, url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("Accept-Language", AccessLanguage)
	if requireAuth {
		if viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME") == "" {
			return nil, errors.New("ISE client username is not provided")
		}
		if viper.GetString("PXGRID_CLIENT_ACCOUNT_PASSWORD") == "" {
			return nil, errors.New("ISE client password is not provided")

		}
		req.SetBasicAuth(viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME"), viper.GetString("PXGRID_CLIENT_ACCOUNT_PASSWORD"))
		return c.client.Do(req)
	}
	return c.client.Do(req)
}

// ReadSessions Read session events from PxGrid
func (c *Controller) ReadSessions(secret, url string, requestBody interface{}) (*http.Response, error) {
	var req *http.Request
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeoutValue*time.Second)
	defer cancel()
	if requestBody != nil {
		requestBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBytes))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("Accept-Language", AccessLanguage)
	if viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME") == "" {
		return nil, errors.New("ISE client username is not provided")
	}
	req.SetBasicAuth(viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME"), secret)
	return c.client.Do(req)

}
