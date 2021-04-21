package lib

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
}

// NewConfig create a new config file
func NewConfig() *Config {
	return &Config{}
}

// GetTLSConfig generate TLS Config
func (c *Config) GetTLSConfig() (*tls.Config, error) {
	endpoint := fmt.Sprintf("%s:%d", viper.GetString("PXGRID_HOST_ADDRESS"), viper.GetInt("ISE_PORT"))
	caCert, err := ExtractServerCert(endpoint)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}, nil
}
