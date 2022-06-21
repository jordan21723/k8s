package utils

/*
here we validate all the data from api ensure it`s fit certain scenario
package api_server

/*
here we validate all the data from api ensure it`s fit certain scenario
*/

import (
	"errors"
	"fmt"
	"k8s-installer/pkg/coredns"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

const (
	dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
)

var (
	namespaceReg    *regexp.Regexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")
	storageClassReg *regexp.Regexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")
)
var v *validator.Validate

func init() {
	v = validator.New()
	v.RegisterValidation("k8s_namespace", validateKubenetesNamespace)
	v.RegisterValidation("k8s_storage_class", validateKubenetesStorageClass)
	v.RegisterValidation("storage_size", validateStorageSize)
}

func ValidateUpgrade(version schema.UpgradableVersion) error {
	err := v.Struct(version)
	if err != nil {
		return err
	}
	reg := regexp.MustCompile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)
	if !reg.MatchString(version.Name) {
		return errors.New(fmt.Sprintf("Upgrade version %s is not valid. Valid version format: [major].[sub].[min]. Example: 1.18.10", version.Name))
	}

	if version.Centos != nil {
		if len(version.Centos.OSVersions) > 0 {
			for osVerId, osVer := range version.Centos.OSVersions {
				if err := checkCentosVersionCheck(osVerId); err != nil {
					return printUpgradeCheckFailedMessage(version.Name, err)
				}
				if len(osVer.X8664.RPMS) > 0 {
					for _, rpmName := range osVer.X8664.RPMS {
						if err := v.Struct(rpmName); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
						if err := checkCentosRPMName(rpmName.Name); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
					}
				}
				if len(osVer.Aarch64.RPMS) > 0 {
					for _, rpmName := range osVer.Aarch64.RPMS {
						if err := v.Struct(rpmName); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
						if err := checkCentosRPMName(rpmName.Name); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
					}
				}
			}
		}
	}

	if version.Ubuntu != nil {
		if len(version.Ubuntu.OSVersions) > 0 {
			for osVerId, osVer := range version.Ubuntu.OSVersions {
				if err := checkUbuntuVersionCheck(osVerId); err != nil {
					return err
				}
				if len(osVer.X8664.Debs) > 0 {
					for _, pkgName := range osVer.X8664.Debs {
						if err := v.Struct(pkgName); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
						if err := checkUbuntuRPMName(pkgName.Name); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
					}
				}
				if len(osVer.Aarch64.Debs) > 0 {
					for _, pkgName := range osVer.Aarch64.Debs {
						if err := v.Struct(pkgName); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
						if err := checkUbuntuRPMName(pkgName.Name); err != nil {
							return printUpgradeCheckFailedMessage(version.Name, err)
						}
					}
				}
			}

		}
	}
	return nil
}

func printUpgradeCheckFailedMessage(version string, err error) error {
	return errors.New(fmt.Sprintf("Validation check failed for k8s Version %s due to %s", version, err.Error()))
}

func checkCentosVersionCheck(ver string) error {
	versionAvailable := map[string]byte{
		"7": 0,
	}
	if _, exists := versionAvailable[ver]; !exists {
		return errors.New(fmt.Sprintf("Centos os version %s is not support", ver))
	}
	return nil
}

func checkUbuntuVersionCheck(ver string) error {
	return nil
}

func checkCentosRPMName(packageName string) error {
	return nil
}

func checkUbuntuRPMName(packageName string) error {
	return nil
}

func DnsValidate(s interface{}) error {
	err := v.Struct(s)
	if err != nil {
		return err
	}

	if resolve, ok := s.(coredns.DomainResolve); ok {
		if resolve.RecordType == constants.DnsRecordTypeA {
			if err := v.Var(resolve.IpAddress, "ipv4"); err != nil {
				return fmt.Errorf("When dns record type set to %s, IpAddress must set to a valid ipv4 address ", constants.DnsRecordTypeA)
			}
		} else if resolve.RecordType == constants.DnsRecordTypeAAAA {
			if err := v.Var(resolve.IpAddress, "ipv6"); err != nil {
				return fmt.Errorf("When dns record type set to %s, IpAddress must set to a valid ipv6 address ", constants.DnsRecordTypeAAAA)
			}
		} else {
			return fmt.Errorf("%s is not a support dns record type", resolve.RecordType)
		}
	}

	return nil
}

func Validate(s interface{}) error {

	err := v.Struct(s)
	if err != nil {
		return err
	}
	if clu, ok := s.(schema.Cluster); ok {
		if len(clu.ExternalLB.NodeIds) > 0 {
			err = exLBValidate(&clu.ExternalLB)
			if err != nil {
				return err
			}
		}
		if err := virtualKubeletValidate(clu); err != nil {
			return err
		}
		if clu.ClusterDnsUpstream != nil {
			for _, address := range clu.ClusterDnsUpstream.Addresses {
				err = v.Var(address, "ipv4")
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func virtualKubeletValidate(cluster schema.Cluster) error {
	for _, master := range cluster.Masters {
		if master.UseVirtualKubelet {
			return errors.New("Cannot use virtual kubelet on master node! ")
		}
	}
	return nil
}

func createVKCheckErrorMsg(field string) error {
	return errors.New(fmt.Sprintf("%s is required when UseVirtualKubelet is set to yes", field))
}

func CalicoValidate(clu *schema.Cluster) error {
	var err error

	if clu.CNI.Calico.EnableDualStack == true {
		err = v.Var(clu.ControlPlane.ServiceV6CIDR, "cidrv6")
		if err != nil {
			return err
		}
	}
	if clu.CNI.Calico.IPAutoDetection != "" {
		if err = ipAutodetection(clu.CNI.Calico.IPAutoDetection, false); err != nil {
			return err
		}
	}

	if clu.CNI.Calico.EnableDualStack {
		if clu.CNI.Calico.IP6AutoDetection != "" {
			if err = ipAutodetection(clu.CNI.Calico.IP6AutoDetection, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func ipAutodetection(val string, isIPV6 bool) error {
	if val == "first-found" {
		return nil
	}

	arr := strings.Split(val, "=")
	if len(arr) != 2 {
		return errors.New("Calico IP Autodetection Method Format Lenght Error")
	}

	//if arr[0] != "can-reach" && arr[0] != "interface" && arr[0] != "skip-interface" && arr[0] != "cidr" {
	if arr[0] != "can-reach" && arr[0] != "interface" && arr[0] != "skip-interface" && arr[0] != "first-found" {
		return errors.New(fmt.Sprintf("Calico IP Autodetection Method Format Type Error, %s is not a support type", arr[0]))
	}

	if arr[0] == "can-reach" {
		var errIP error
		if isIPV6 {
			errIP = v.Var(arr[1], "ipv6")
		} else {
			errIP = v.Var(arr[1], "ipv4")
		}
		//errUrl = v.Var(arr[1], "dns")
		isDomain := domainValidate(arr[1])
		if errIP != nil && !isDomain {
			if isIPV6 {
				return errors.New(fmt.Sprintf("can-reach must be set to either url or ipv6 address but got %s", arr[1]))
			} else {
				return errors.New(fmt.Sprintf("can-reach must be set to either url or ipv4 address but got %s", arr[1]))
			}
		}
		return nil
	}

	// so far cidr method is not support by our calico image version
	// consider uncomment the following block when it does
	/*	if arr[0] == "cidr" {
		_,_,err := net.ParseCIDR(arr[1])
		if err != nil {
			return errors.New(fmt.Sprintf("cidr must be set to a valid CIDR, got %s", arr[1]))
		}
	}*/
	return nil
}

func domainValidate(domain string) bool {
	// since govalidate doest not work well on validate domains
	// we have to do it own
	RegExp := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z
 ]{2,3})$`)
	return RegExp.MatchString(domain)
}

func exLBValidate(exlb *schema.ExternalLB) error {
	return v.Var(exlb.ClusterVipV4, "ipv4")
}

func validateKubenetesNamespace(fl validator.FieldLevel) bool {
	namespace := fl.Field().String()
	return namespaceReg.MatchString(namespace)
}

func validateKubenetesStorageClass(fl validator.FieldLevel) bool {
	storageClass := fl.Field().String()
	return storageClassReg.MatchString(storageClass)
}

func validateStorageSize(fl validator.FieldLevel) bool {
	num := fl.Field().Int()
	min := constants.StorageMinSize
	max := constants.StorageMaxSize
	if constants.Size(num) < min {
		return false
	}
	if constants.Size(num) > max {
		return false
	}
	return true
}

func RemoveDuplicateNode(cluster schema.Cluster) []schema.ClusterNode {
	reduced := map[string]byte{}
	result := []schema.ClusterNode{}
	for _, master := range cluster.Masters {
		reduced[master.NodeId] = 1
		result = append(result, master)
	}

	for _, worker := range cluster.Workers {
		if _, exits := reduced[worker.NodeId]; !exits {
			result = append(result, worker)
		}
	}

	return result
}
