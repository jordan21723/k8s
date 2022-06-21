package constants

const (
	ChallengeCodeListCluster uint64 = 1 << iota // role function is malformed
	ChallengeCodeManageRegion
	ChallengeCodeCreateCluster
	ChallengeCodeDeleteCluster
	ChallengeCodeAddNodeToCluster
	ChallengeCodeRemoveNodeFromCluster
	ChallengeCodeListNode
	ChallengeCodeManageDNS
	ChallengeCodeEnableOrDisableNode
	ChallengeCodeManageUser
	ChallengeCodeBackupAndRestore
	ChallengeCodeKubectlExec
	ChallengeCodeDeleteUser
	ChallengeCodeManageRole
	ChallengeCodeViewRoleDetail
	ChallengeCodeCreateRole
	ChallengeCodeDeleteRole
	ChallengeCodeContinueTask
	ChallengeCodeListNodeTask
	ChallengeCodeGetNodeTask
	ChallengeListUpgradeVersion
	ChallengeConfigUpgradeVersion
	ChallengeUpgradeCluster
	ChallengeCodeConfigPlatform
	ChallengeCodeManageCert
)