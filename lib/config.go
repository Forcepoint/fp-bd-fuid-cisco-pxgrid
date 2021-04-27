package lib

import (
	"crypto/tls"
	"crypto/x509"
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
	caCert, err := ExtractServerCert(viper.GetString("PXGRID_HOST_ADDRESS"), viper.GetInt("ISE_PORT"))
	if err != nil {
		return nil, err
	}
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
