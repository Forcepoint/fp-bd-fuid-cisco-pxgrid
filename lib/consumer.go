package lib

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"time"
)

type ReadSessionInput struct {
	StartTimestamp *time.Time `json:"startTimestamp"`
}

// ValidateUsernamePassword ensure the yaml config file contains ISE Credentials
func ValidateUsernamePassword() error {
	if viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME") == "" {
		return errors.New("Ise client username is not provided")
	}
	if viper.GetString("PXGRID_CLIENT_ACCOUNT_PASSWORD") == "" {
		return errors.New("Ise client password is not provided")
	}
	return nil
}

// GetSessionRestUrl extract the Session REST API URL from a service
func GetSessionRestUrl(services []Services) (string, string, error) {
	for _, s := range services {
		if s.Properties.SessionTopic != "" && s.Properties.RestBaseUrl != "" && s.NodeName != "" {
			return s.Properties.RestBaseUrl, s.NodeName, nil
		}
	}
	return "", "", errors.New("cannot find any restBaseUrl for sessions in any service")

}

// SessionListener listen to session events
func SessionListener(secret, restUrl, timeStampFilePath string, controller *Controller, fuidController *FUIDController, displayProcess bool) error {
	restUrl = fmt.Sprintf("%s/%s", restUrl, GetSessionEndpoint)
	readSessionInput, err := GetLatestSessionTimeStamp(timeStampFilePath)
	if err != nil {
		return err
	}
	resp, err := controller.ReadSessions(secret, restUrl, &readSessionInput)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s  %s", resp.StatusCode, resp.Status, "not authorized"))
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		respBodyString := string(respBody)
		if respBodyString != "" {
			return errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s, Body: %s", resp.StatusCode, resp.Status, respBodyString))
		}
		return errors.New(fmt.Sprintf("UnexpectedResponseError: status_code: %d, statusReason: %s", resp.StatusCode, resp.Status))
	}
	var sessions IseSessions
	respBody, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &sessions); err != nil {
		if err := FixJson(respBody, &sessions); err != nil {
			return errors.Wrap(err, "SessionListener")
		}
	}
	if len(sessions.Sessions) != 0 {
		if displayProcess {
			logrus.Infof("Latest stored timestamp: %s", readSessionInput.StartTimestamp)
			logrus.Infof("Number of new session events: %d", len(sessions.Sessions))
		}
		if err := ProcessSessions(&sessions, timeStampFilePath, fuidController, displayProcess); err != nil {
			return err
		}
	}
	return nil
}

// ProcessSessions process list of session events
func ProcessSessions(sessions *IseSessions, timeStampFilePath string, fuidController *FUIDController, displayProcess bool) error {
	latestTimeStamp, err := readTimeStampFromDisk(timeStampFilePath)
	if err != nil {
		return err
	}
	maxTimeStamp := latestTimeStamp.StartTimestamp
	for _, sess := range sessions.Sessions {
		if sess.Timestamp.After(*latestTimeStamp.StartTimestamp) && !sess.Timestamp.Equal(*latestTimeStamp.StartTimestamp) {
			if sess.Timestamp.After(*maxTimeStamp) && !sess.Timestamp.Equal(*maxTimeStamp) {
				maxTimeStamp = sess.Timestamp
			}
			if err := fuidController.UserManager(&sess, displayProcess); err != nil {
				return err
			}
		}
	}
	if maxTimeStamp.After(*latestTimeStamp.StartTimestamp) && !maxTimeStamp.Equal(*latestTimeStamp.StartTimestamp) {
		if err := saveTimeStampToDisk(maxTimeStamp, timeStampFilePath); err != nil {
			return err
		}
		if displayProcess {
			logrus.Infof("New latest timestamp (%s) has been written to disk", maxTimeStamp.String())
		}
	}
	return nil
}

// saveTimeStampToDisk store th timestamp.
func saveTimeStampToDisk(newTimestamp *time.Time, timeStampFilePath string) error {
	newTimestampPlus := newTimestamp.Add(time.Millisecond)
	t := &ReadSessionInput{StartTimestamp: &newTimestampPlus}
	timeStampBytes, err := json.Marshal(t)
	if err != nil {
		return errors.New(fmt.Sprintf("saveTimeStampToDisk %s", err.Error()))
	}
	if err := ioutil.WriteFile(timeStampFilePath, timeStampBytes, 0666); err != nil {
		return errors.New(fmt.Sprintf("saveTimeStampToDisk %s", err.Error()))
	}
	return nil
}

// readTimeStampFromDisk read the timestamp from the disk
func readTimeStampFromDisk(timeStampFilePath string) (*ReadSessionInput, error) {
	var latestTimestamp ReadSessionInput
	timeStampByte, err := ioutil.ReadFile(timeStampFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("saveTimeStampToDisk %s", err.Error()))
	}
	if err := json.Unmarshal(timeStampByte, &latestTimestamp); err != nil {
		return nil, errors.New(fmt.Sprintf("saveTimeStampToDisk %s", err.Error()))
	}
	if latestTimestamp.StartTimestamp == nil {
		return nil, errors.New("StartTimestamp is nil in the stored latestTimestamp")
	}
	return &latestTimestamp, nil
}

// GetLatestSessionTimeStamp get the timestamp for the latest processed session
func GetLatestSessionTimeStamp(timeStampFilePAth string) (*ReadSessionInput, error) {
	if !IsFileExist(timeStampFilePAth) {
		t := time.Now()
		t = t.Add(-6 * time.Hour)
		latestTimestamp := &ReadSessionInput{StartTimestamp: &t}
		timeStampBytes, err := json.Marshal(latestTimestamp)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("GetLatestSessionTimeStamp %s", err.Error()))
		}
		if err := ioutil.WriteFile(timeStampFilePAth, timeStampBytes, 0666); err != nil {
			return nil, errors.New(fmt.Sprintf("GetLatestSessionTimeStamp %s", err.Error()))
		}
		return latestTimestamp, nil
	}
	var latestTimestamp ReadSessionInput
	timeStampByte, err := ioutil.ReadFile(timeStampFilePAth)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetLatestSessionTimeStamp %s", err.Error()))
	}
	if err := json.Unmarshal(timeStampByte, &latestTimestamp); err != nil {
		return nil, errors.New(fmt.Sprintf("GetLatestSessionTimeStamp %s", err.Error()))
	}
	return &latestTimestamp, nil
}
