package lib

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type FUIDUser struct {
	Dn             string   `json:"dn,omitempty"`
	ChangeType     string   `json:"changetype,omitempty"`
	SAMAccountName string   `json:"sAMAccountName,omitempty"`
	NTLMIdentity   string   `json:"NTLMIdentity,omitempty"`
	Mail           string   `json:"mail,omitempty"`
	Ipv4Addresses  []string `json:"ipv4_addresses,omitempty"`
	Ipv6Addresses  []string `json:"ipv6_addresses,omitempty"`
	ObjectGUID     string   `json:"objectGUID,omitempty"`
	Groups         []string `json:"groups,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
}

type AllUsers struct {
	Users []FUIDUser `json:"users"`
}

type FUIDController struct {
	client *http.Client
}

// GetTLSConfig Get TLS Config for FUID API
func (f *FUIDController) GetTLSConfig() (*tls.Config, error) {
	caCert, err := ExtractServerCert(viper.GetString("FUID_IP_ADDRESS"), viper.GetInt("FUID_PORT"))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}
	return tlsConfig, nil
}

// NewFUIDController Create a Controller for FUID API
func NewFUIDController() (*FUIDController, error) {
	controller := FUIDController{}
	tlsConfig, err := controller.GetTLSConfig()
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	controller.client = &http.Client{Transport: transport}
	return &controller, nil
}

// GetUser Search for a specific use in FUID Database
func (f *FUIDController) GetUser(userNTLMIdentity string) (*FUIDUser, error) {
	endpoint := fmt.Sprintf("%s/%s", UserNtlmIdentityEndpoint, userNTLMIdentity)
	resp, err := f.SendRequest(endpoint, "", nil, http.MethodGet)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed in reading user %s from FUID with status error %d %s", userNTLMIdentity, resp.StatusCode, resp.Status)
	}
	if resp == nil {
		return nil, errors.Errorf("recived nil response for reading user %s from FUID with status error %d %s", userNTLMIdentity, resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	var user FUIDUser
	if err := json.Unmarshal(responseBody, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SendRequest send a request to FUID API
func (f *FUIDController) SendRequest(endPoint, parameters string, requestBody interface{}, requestMethod string) (*http.Response, error) {
	var req *http.Request
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeoutValue*time.Second)
	defer cancel()
	requestUrl, err := generateUrl(endPoint, parameters)
	if err != nil {
		return nil, err
	}
	urlParsed, err := url.Parse(requestUrl)
	if err != nil {
		return nil, err
	}
	requestUrl = urlParsed.String()
	if requestBody != nil {
		requestBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, requestMethod, requestUrl, bytes.NewBuffer(requestBytes))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, requestMethod, requestUrl, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("Accept-Language", AccessLanguage)
	if requestMethod != http.MethodGet {
		if viper.GetString("FUID_API_USERNAME") != "" && viper.GetString("FUID_API_PASSWORD") != "" {
			req.SetBasicAuth(viper.GetString("FUID_API_USERNAME"), viper.GetString("FUID_API_PASSWORD"))
		}
	}
	resp, err := f.client.Do(req)
	if resp == nil {
		return nil, err
	}
	return resp, err
}

// UserManager manager a session, if your is not exists in FUID database, create it, otherwise update the user IP Addresses ang Groups
func (f *FUIDController) UserManager(sess *Sessions, displayProcess bool) error {
	userAccountName := sess.AdUserSamAccountName
	useNetBiosName := sess.AdUserNetBiosName
	username := fmt.Sprintf("%s\\%s", useNetBiosName, userAccountName)
	user, err := f.GetUser(username)
	if err != nil {
		if err == NotFound {
			//connect to AD and read the user Object
			logrus.Warningf("User %s is not exist in FUID Database", sess.AdUserSamAccountName)
			ldapConnector, err := NewADConnector()
			if err != nil {
				return err
			}
			if displayProcess {
				logrus.Infof("Connecting with AD Domain Conttroler %s", viper.GetString("AD_LDAP_HOST"))
			}
			defer ldapConnector.Close()
			userEntity, err := GetLdapElement(sess.AdUserSamAccountName, ldapConnector)
			if err != nil {
				return err
			}
			if displayProcess {
				logrus.Infof("Read User Object from AD Domanin Controller for user %s", sess.AdUserSamAccountName)
			}
			err = f.PostUser(userEntity, sess, displayProcess)
			if err != nil {
				return err
			}

			return nil
		} else {
			return err
		}
	}
	if displayProcess {
		logrus.Infof("Succefully read user object from FUID LDAP for user %s", user.NTLMIdentity)
	}
	if err := f.PutUser(user, sess, displayProcess); err != nil {
		return err
	}
	return nil
}

// PutUser Update a user's IP addresses and Groups
func (f *FUIDController) PutUser(user *FUIDUser, sess *Sessions, displayProcess bool) error {
	switch sess.State {
	case AUTHENTICATED:
		if displayProcess {
			logrus.Infof("User %s has authenticated", user.NTLMIdentity)
		}
		changeType := ChangeTypeAdd
		if user.Ipv4Addresses != nil || len(user.Ipv4Addresses) != 0 {
			changeType = ChangeTypeModify
		}
		var newUser FUIDUser
		newUser.ObjectGUID = user.ObjectGUID
		newUser.ChangeType = changeType
		newUser.Ipv4Addresses = sess.IpAddresses
		endpoint := fmt.Sprintf("%s/%s", UserEndpoint, newUser.ObjectGUID)
		resp, err := f.SendRequest(endpoint, "", &newUser, http.MethodPut)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New("Not Authorized to do Post request to FUID API")
		}
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("add user to FUID statusCode %d %s", resp.StatusCode, resp.Status)
		}
		if displayProcess {
			logrus.Infof("%s IP addresses for user %s", changeType, user.NTLMIdentity)
		}
		return nil

	case DISCONNECTED:
		if displayProcess {
			logrus.Infof("User %s has Disconnected(session termination)", user.NTLMIdentity)
		}
		var newUser FUIDUser
		newUser.ObjectGUID = user.ObjectGUID
		newUser.ChangeType = ChangeTypeDelete
		newUser.Ipv4Addresses = sess.IpAddresses
		endpoint := fmt.Sprintf("%s/%s", UserEndpoint, newUser.ObjectGUID)
		resp, err := f.SendRequest(endpoint, "", &newUser, http.MethodPut)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New("Not Authorized to do Post request to FUID API")
		}
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("add user to FUID statusCode %d %s", resp.StatusCode, resp.Status)
		}
		if displayProcess {
			logrus.Infof("delete IP addresses for user  %s", user.NTLMIdentity)
		}
		return nil
	}
	return nil

}

// PostUser Create a user in FUID Database.
func (f *FUIDController) PostUser(userEntity *LdapElement, sess *Sessions, displayProcess bool) error {
	var newUser FUIDUser
	nTm := fmt.Sprintf("%s\\%s", sess.AdUserNetBiosName, sess.AdUserSamAccountName)
	newUser.NTLMIdentity = nTm
	newUser.Dn = sess.AdUserResolvedDns
	newUser.Ipv4Addresses = sess.IpAddresses
	newUser.SAMAccountName = sess.AdUserSamAccountName
	newUser.ObjectGUID = userEntity.Attributes.ObjectGUID
	newUser.Groups = userEntity.Attributes.MemberOf
	endpoint := fmt.Sprintf("%s/%s", UserEndpoint, userEntity.Attributes.ObjectGUID)
	resp, err := f.SendRequest(endpoint, "", &newUser, http.MethodPost)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusBadRequest {
		defer resp.Body.Close()
		d, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("error in posting user to FUID. %s", string(d))
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("Not Authorized to do Post request to FUID API")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("add user to FUID statusCode %d %s", resp.StatusCode, resp.Status)
	}
	if displayProcess {
		logrus.Infof("use %s has been written to FUILD Database", sess.AdUserSamAccountName)
	}
	return nil
}

// generateUrl Generate FUID endpoint URl.
func generateUrl(endpoint, parameters string) (string, error) {
	if viper.GetString("FUID_IP_ADDRESS") == "" {
		return "", errors.New("The FUID API IP address is not provided")
	}
	if viper.GetInt("FUID_PORT") == 0 {
		return "", errors.New("The FUID API port number is not provided")
	}
	generatedUrl := fmt.Sprintf("https://%s:%d/api/uid/v1.0/%s", viper.GetString("FUID_IP_ADDRESS"), viper.GetInt("FUID_PORT"), endpoint)
	if parameters != "" {
		generatedUrl = fmt.Sprintf("%s?%s", generatedUrl, parameters)
	}
	return generatedUrl, nil
}
