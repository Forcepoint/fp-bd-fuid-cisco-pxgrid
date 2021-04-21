package lib

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	NotFound error = errors.New("Not Found")
)

const (
	RequestTimeoutValue           = 5 //the timeout in seconds for ISE requests
	AccessLanguage                = "application/json"
	ContentType                   = "application/json"
	PxGridCreateClientEndPoint    = "pxgrid/control/AccountCreate"
	PxGridAccountActivateEndPoint = "pxgrid/control/AccountActivate"
	ServiceLookup                 = "pxgrid/control/ServiceLookup"
	ServiceLookupSessions         = "com.cisco.ise.session"
	AccessSecretEndpoint          = "pxgrid/control/AccessSecret"
	NoServiceAvailable            = "no service available"
	Enabled                       = "ENABLED"
	GetSessionEndpoint            = "getSessions"
	//FUID
	UserNtlmIdentityEndpoint = "user/ntlm-identity"
	UserEndpoint             = "user"
	FuidAllUsers             = "users"
	AUTHENTICATED            = "AUTHENTICATED"
	AUTHENTICATING           = "AUTHENTICATING"
	POSTURED                 = "POSTURED"
	DISCONNECTED             = "DISCONNECTED"
	ChangeTypeAdd            = "add"
	ChangeTypeModify         = "modify"
	ChangeTypeDelete         = "delete"
	//LDAP
	LdapFilerFormat = "(&(objectCategory=person)(objectClass=user)(sAMAccountName=%s))"
	LdapAttributes  = "memberOf,objectclass,objectGUID,sAMAccountName,userPrincipalName,CN"
)

func GetEndpointUrl(endpointName string) string {
	return fmt.Sprintf("https://%s:%d/%s", viper.GetString("PXGRID_HOST_ADDRESS"),
		viper.GetInt("ISE_PORT"), endpointName)
}
