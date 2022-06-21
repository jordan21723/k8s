package schema

import "k8s-installer/pkg/dep"

type UpgradePlan struct {
	Id                string                  `json:"id"`
	IsPackageReady    bool                    `json:"is_rmp_ready"`
	ClusterId         string                  `json:"cluster_id"`
	NodeChangePlan    []UpgradeNodeCheckState `json:"rmp_check_status"`
	ImageChangePlan   []ImageCheckState       `json:"image_change_plan"`
	IsImageReady      bool                    `json:"is_image_ready"`
	Plan              []string                `json:"plan"`
	MessagePlanResult string                  `json:"message_plan_result"`
	Status            string                  `json:"status"`
	TargetVersion     string                  `json:"target_version"`
	OperationId       string                  `json:"operation_id"`
	Operations        []string                `json:"operations"`
	ApplyDate         string                  `json:"apply_date"`
}

type UpgradeNodeCheckState struct {
	NodeID             string              `json:"node_id"`
	OSFamily           string              `json:"os_family"`
	CpuArch            string              `json:"cpu_arch"`
	NodeIPV4           string              `json:"node_ipv_4"`
	PackageCheckResult []PackageCheckState `json:"package_check_result"`
}

type PackageCheckState struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ImageCheckState struct {
	Name         string `json:"name"`
	CurrentTag   string `json:"old_version"`
	UpgradeToTag string `json:"change_to"`
	Status       string `json:"status"`
}

type eCheckState struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (UpgradableVersion UpgradableVersion) ToDep() dep.DepMap {
	deps := dep.DepMap{}
	if UpgradableVersion.Centos != nil {
		if len(UpgradableVersion.Centos.OSVersions) > 0 {
			initDeps(deps, "centos", UpgradableVersion.Name)
			for _, packageSets := range UpgradableVersion.Centos.OSVersions {
				for _, packageName := range packageSets.X8664.RPMS {
					deps["centos"][UpgradableVersion.Name]["x86_64"][packageName.Name] = packageName.Name
				}
				for _, packageName := range packageSets.Aarch64.RPMS {
					deps["centos"][UpgradableVersion.Name]["aarch64"][packageName.Name] = packageName.Name
				}
			}
		}
	}
	if UpgradableVersion.Ubuntu != nil {
		if len(UpgradableVersion.Ubuntu.OSVersions) > 0 {
			initDeps(deps, "ubuntu", UpgradableVersion.Name)
			for _, packageSets := range UpgradableVersion.Centos.OSVersions {
				for _, packageName := range packageSets.X8664.RPMS {
					deps["ubuntu"][UpgradableVersion.Name]["x86_64"][packageName.Name] = packageName.Name
				}
				for _, packageName := range packageSets.Aarch64.RPMS {
					deps["ubuntu"][UpgradableVersion.Name]["aarch64"][packageName.Name] = packageName.Name
				}
			}
		}
	}
	return deps
}

func initDeps(deps dep.DepMap, osFamily, upgradeVer string) {
	deps[osFamily] = dep.DepVersion{}
	deps[osFamily][upgradeVer] = dep.DepArch{}
	deps[osFamily][upgradeVer]["x86_64"] = dep.DepPackage{}
	deps[osFamily][upgradeVer]["aarch64"] = dep.DepPackage{}
}

type UpgradableVersion struct {
	Name           string          `json:"name" validate:"required"`
	VersionState   string          `json:"version_state" validate:"oneof=alpha beta GA"`
	Description    string          `json:"description"`
	Created        string          `json:"date_created,omitempty"`
	Modified       string          `json:"date_modified,omitempty"`
	CreatedBy      string          `json:"created_by"`
	LastModifiedBy string          `json:"last_modified_by"`
	Centos         *OSFamilyCentos `json:"centos,omitempty"`
	Ubuntu         *OSFamilyUbuntu `json:"ubuntu,omitempty"`
}

type OSFamilyCentos struct {
	OSVersions map[string]OSFamilyCentosVersion `json:"os_versions,omitempty" validate:"required"`
}

type OSFamilyUbuntu struct {
	OSVersions map[string]OSFamilyUbuntuVersion `json:"os_versions,omitempty" validate:"required"`
}

type OSFamilyCentosVersion struct {
	X8664   OSFamilyCentosCpuArchX8664   `json:"x86_64"`
	Aarch64 OSFamilyCentosCpuArchAarch64 `json:"aarch_64"`
}

type OSFamilyUbuntuVersion struct {
	X8664   OSFamilyUbuntuCpuArchX8664   `json:"x86_64,omitempty"`
	Aarch64 OSFamilyUbuntuCpuArchAarch64 `json:"aarch_64,omitempty"`
}

type OSFamilyCentosCpuArchX8664 struct {
	RPMS []OSPackage `json:"rpms,omitempty"`
}

type OSFamilyCentosCpuArchAarch64 struct {
	RPMS []OSPackage `json:"rpms,omitempty"`
}

type OSFamilyUbuntuCpuArchX8664 struct {
	Debs []OSPackage `json:"debs,omitempty"`
}

type OSFamilyUbuntuCpuArchAarch64 struct {
	Debs []OSPackage `json:"debs,omitempty"`
}

type OSPackage struct {
	Name string `json:"name,omitempty" validate:"required"`
}
