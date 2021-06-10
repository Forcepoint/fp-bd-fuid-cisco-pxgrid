package lib

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/ldap.v2"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type KeyValue map[string]interface{}

type LdapEntity struct {
	DN         string
	Attributes []KeyValue
}

// LdapElement  holds the DN and Attributes of an LDAP/AD entry.
type LdapElement struct {
	DN         string `json:"DN"`
	Attributes Attributes
}

type Attributes struct {
	Cn             string   `json:"cn"`
	MemberOf       []string `json:"memberOf"`
	ObjectGUID     string   `json:"objectGUID"`
	SAMAccountName string   `json:"sAMAccountName"`
}

func NewADConnector() (*ldap.Conn, error) {
	if viper.GetString("AD_LDAP_USER_DN") == "" {
		return nil, errors.New("AD LDAP username is not provided")
	}
	ldapUsername := viper.GetString("AD_LDAP_USER_DN")
	if viper.GetString("AD_LDAP_HOST") == "" {
		return nil, errors.New("AD Domain Controller Ip address is not provided")
	}
	if viper.GetString("AD_LDAP_PASSWORD") == "" {
		return nil, errors.New("AD Domain Controller admin password is not provided")
	}
	ldapConnector, err := connectToDirectoryServerTLS(viper.GetString("AD_LDAP_HOST"), viper.GetInt("AD_PORT"),
		ldapUsername, viper.GetString("AD_LDAP_PASSWORD"), viper.GetInt("LDAP_TIMEOUT"), true, "", "")
	if err != nil {
		return nil, err
	}
	return ldapConnector, nil

}

// connect to LDAP
func connectToDirectoryServer(Host string, Port int, Username, Password string, ConnTimeout int) (*ldap.Conn, error) {
	ldap.DefaultTimeout = time.Duration(ConnTimeout) * time.Second
	ldapConnector, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", Host, Port))
	if err != nil {
		return nil, err
	}
	err = ldapConnector.Bind(Username, Password)
	if err != nil {
		return nil, err
	}
	return ldapConnector, nil
}

// connect to LDAPs
func connectToDirectoryServerTLS(Host string, Port int, Username, Password string, ConnTimeout int, CRTInsecureSkipVerify bool,
	CRTValidFor, CRTPath string) (*ldap.Conn, error) {
	ldap.DefaultTimeout = time.Duration(ConnTimeout) * time.Second
	tlsConfig := new(tls.Config)
	var err error
	if !CRTInsecureSkipVerify {
		tlsConfig, err = tlsConfigNoSkipVerify(CRTInsecureSkipVerify, CRTValidFor, CRTPath)
		if err != nil {
			return nil, err
		}
	} else {
		tlsConfig = tlsConfigSkipVerify(CRTInsecureSkipVerify)
	}
	ldapConnector, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", Host, Port), tlsConfig)
	if err != nil {
		if strings.Contains(err.Error(), "connection reset by peer") {
			return nil, errors.Errorf("Failed in dialing LDAP with TLS. ensure LDAP server is configured to use SSL over port 636")
		}
		return nil, err
	}
	err = ldapConnector.Bind(Username, Password)
	if err != nil {
		return nil, err
	}
	return ldapConnector, nil
}

// skip TLS verification
func tlsConfigSkipVerify(CRTInsecureSkipVerify bool) *tls.Config {
	tlsConfig := new(tls.Config)
	tlsConfig.InsecureSkipVerify = CRTInsecureSkipVerify
	return tlsConfig
}

// verify TLS cert
func tlsConfigNoSkipVerify(CRTInsecureSkipVerify bool, CRTValidFor, CRTPath string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	caCert, err := ioutil.ReadFile(CRTPath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	tlsConfig.InsecureSkipVerify = CRTInsecureSkipVerify
	tlsConfig.ServerName = CRTValidFor
	tlsConfig.RootCAs = pool
	return tlsConfig, nil
}

func getFromLDAP(connect *ldap.Conn, LDAPBaseDN, LDAPFilter string, LDAPAttribute []string, LDAPPage uint32) ([]LdapEntity, error) {
	searchRequest := ldap.NewSearchRequest(LDAPBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, LDAPFilter, LDAPAttribute, nil)
	sr, err := connect.SearchWithPaging(searchRequest, LDAPPage)
	if err != nil {
		return nil, err
	}
	var ADElements []LdapEntity
	for _, entry := range sr.Entries {
		NewADEntity := new(LdapEntity)
		NewADEntity.DN = entry.DN
		for _, attrib := range entry.Attributes {
			NewADEntity.Attributes = append(NewADEntity.Attributes, KeyValue{attrib.Name: attrib.Values})
		}
		ADElements = append(ADElements, *NewADEntity)
	}
	return ADElements, nil
}

func GetLdapElement(username string, ldapConnector *ldap.Conn) (*LdapElement, error) {
	baseDn, err := generateLdapBaseDn()
	if err != nil {
		return nil, err
	}
	filter := fmt.Sprintf(viper.GetString("LDAP_FILTER"), username)
	attributes := strings.Split(viper.GetString("LDAP_ATTRIBUTES"), ",")
	LDAPElements, err := getFromLDAP(ldapConnector, baseDn, filter, attributes, uint32(viper.GetInt("LDAP_PAGES")))
	if err != nil {
		return nil, err
	}
	if len(LDAPElements) == 0 {
		return nil, errors.Errorf("could not find use %s in LDAP database", username)
	}
	if len(LDAPElements) > 1 {
		return nil, errors.Errorf("multiple user with name: %s found in LDAP database", username)

	}
	user, err := HandleElement(LDAPElements[0])
	if err != nil {
		return nil, err
	}
	return user, nil
}

func HandleElement(element LdapEntity) (*LdapElement, error) {
	var ldapElement LdapElement
	ldapElement.DN = element.DN
	for _, maps := range element.Attributes {
		for key, value := range maps {
			switch key {
			case "cn":
				ldapElement.Attributes.Cn = value.([]string)[0]
			case "memberOf":
				ldapElement.Attributes.MemberOf = value.([]string)
			case "objectGUID":
				s, r := uuid.Parse(fmt.Sprintf("%x", value.([]string)[0][:]))
				if r != nil {
					log.Fatal(r)
				}
				ldapElement.Attributes.ObjectGUID = handleGUID(s.String())
			case "sAMAccountName":
				ldapElement.Attributes.SAMAccountName = value.([]string)[0]
			}
		}
	}
	return &ldapElement, nil

}

func handleGUID(guid string) string {
	parts := strings.Split(guid, "-")
	partOne := []int{6, 7, 4, 5, 2, 3, 0, 1}
	partTwoPositions := []int{2, 3, 0, 1}
	partThreePositions := []int{2, 3, 0, 1}
	partOneFixed := swapPosition(partOne, parts[0])
	partTwoFixed := swapPosition(partTwoPositions, parts[1])
	partThreeFixed := swapPosition(partThreePositions, parts[2])
	parts[0] = partOneFixed
	parts[1] = partTwoFixed
	parts[2] = partThreeFixed
	return strings.Join(parts, "-")

}

func swapPosition(positions []int, data string) string {
	newArr := make([]rune, 0, len(positions))
	r := []rune(data)
	for _, i := range positions {
		newArr = append(newArr, r[i])
	}
	return string(newArr)
}

func generateLdapUserDn(username string) (string, error) {
	baseDn, err := generateLdapBaseDn()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("CN=%s,CN=Users,%s", username, baseDn), nil

}
func generateLdapBaseDn() (string, error) {
	//	Username = "cn=Administrator,cn=Users,dc=iselab,dc=local"
	if viper.GetString("AD_DOMAIN_NAME") == "" {
		return "", errors.New("AD domain-name is not provided")
	}
	parts := strings.Split(viper.GetString("AD_DOMAIN_NAME"), ".")
	value := strings.Join(parts, ",DC=")
	return fmt.Sprintf("DC=%s", value), nil
}
