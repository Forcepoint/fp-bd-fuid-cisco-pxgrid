package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func IsFileExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetController() (*Controller, error) {
	TLSConfig := NewConfig()
	controller, err := NewControl(TLSConfig)
	if err != nil {
		return nil, err
	}
	return controller, nil
}

func ExtractServerCert(host string, port int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := fmt.Sprintf("openssl s_client -connect %s:%d 2> /dev/null | sed -n '/^-----BEGIN CERTIFICATE-----$/,/^-----END CERTIFICATE-----$/p'", host, port)
	cmd := exec.CommandContext(ctx, "bash", "-c", c)
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errors.New(fmt.Sprintf("timeout exceed for extracting the server certificate, ensure the integration host-machine can reach %s", host))
	}
	if err != nil {
		return nil, err
	}
	return out, err
}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}

func FixJson(data []byte, st interface{}) error {
	upTo := len(string(data)) - 1
	for {
		dataString := string(data)
		newString := dataString[:upTo]
		newString = newString + "]}"
		if err := json.Unmarshal([]byte(newString), st); err == nil {
			return nil
		}
		upTo = upTo - 1
		if upTo < 1000 {
			return errors.New("cannot umarshall sessions")
		}
	}
}
