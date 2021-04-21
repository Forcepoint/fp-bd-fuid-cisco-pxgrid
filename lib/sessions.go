package lib

import "time"

type IseSessions struct {
	Sessions []Sessions `json:"sessions"`
}

type Sessions struct {
	Timestamp                *time.Time `json:"timestamp"`
	State                    string     `json:"state"`
	Username                 string     `json:"userName"`
	CallingStationId         string     `json:"callingStationId"`
	IpAddresses              []string   `json:"ipAddresses"`
	MacAddress               string     `json:"macAddress"`
	NasIpAddress             string     `json:"nasIpAddress"`
	NasIdentifier            string     `json:"nasIdentifier"`
	AdNormalizedUser         string     `json:"adNormalizedUser"`
	AdUserDomainName         string     `json:"adUserDomainName"`
	AdUserNetBiosName        string     `json:"adUserNetBiosName"`
	AdUserResolvedIdentities string     `json:"adUserResolvedIdentities"`
	AdUserResolvedDns        string     `json:"adUserResolvedDns"`
	AdUserQualifiedName      string     `json:"adUserQualifiedName"`
	AdUserSamAccountName     string     `json:"adUserSamAccountName"`
	Providers                []string   `json:"providers"`
	EndpointCheckResult      string     `json:"endpointCheckResult"`
	IdentitySourcePortStart  int        `json:"identitySourcePortStart"`
	IdentitySourcePortEnd    int        `json:"identitySourcePortEnd"`
	IdentitySourcePortFirst  int        `json:"identitySourcePortFirst"`
	IsMachineAuthentication  string     `json:"isMachineAuthentication"`
	NetworkDeviceProfileName string     `json:"networkDeviceProfileName"`
	MdmRegistered            bool       `json:"mdmRegistered"`
	MdmCompliant             bool       `json:"mdmCompliant"`
	MdmDiskEncrypted         bool       `json:"mdmDiskEncrypted"`
	MdmJailBroken            bool       `json:"mdmJailBroken"`
	MdmPinLocked             bool       `json:"mdmPinLocked"`
}
