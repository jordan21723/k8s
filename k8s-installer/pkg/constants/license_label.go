package constants

const (
	ErrLicenseInvalid = 10000
)

const (
	LLConsole uint16 = 1 << iota // role function is malformed
	LLMiddlePlatform
	LLNFS
	LLEFK
	LLGAP
	LLOpenStack
	LLPGO
	LLIngress
	LLClusterLB
	LLHelm
	LLKsInstaller
	LLKsClusterConf
	LLHarbor
	LLMinio
	LLVelero
	LLDnsServerDeploy
)

var LicenseLabelList = map[string]uint16{
	"web-console":     LLConsole,
	"middle-platform": LLMiddlePlatform,
	"nfs":             LLNFS,
	"efk":             LLEFK,
	"gap":             LLGAP,
	"minio":           LLMinio,
	"velero":          LLVelero,
	"openstack":       LLOpenStack,
	"pgo":             LLPGO,
	"ingress":         LLIngress,
	"cluster-lb":      LLClusterLB,
	"helm":            LLHelm,
	"ks-installer":    LLKsInstaller,
	"ks-clusterconf":  LLKsClusterConf,
	"DnsServerDeploy": LLDnsServerDeploy,
}
