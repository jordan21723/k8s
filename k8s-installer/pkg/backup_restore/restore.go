package backup_restore

type RestoreCmd struct {
	Action      string
	BackupName  string
	RestoreName string
	Args        []string
}

func NewRestoreCmd(action, backupName, restoreName string, args []string) *RestoreCmd {
	return &RestoreCmd{
		Action:      action,
		BackupName:  backupName,
		RestoreName: restoreName,
		Args:        args,
	}
}

func (v *RestoreCmd) ActionCmd() []string {
	switch v.Action {
	case Create:
		return v.clusterRestoreCreate(setName(Restore), v.BackupName)
	case Delete:
		return v.clusterRestoreDelete(v.RestoreName)
	case List:
		return v.clusterRestoreList()
	case Describe:
		return v.clusterRestoreDescribe(v.RestoreName)
	default:
		return []string{}
	}
}

func (v *RestoreCmd) clusterRestoreCreate(restoreName, backupName string) []string {
	return veleroCmdStr(Create, Restore, restoreName,
		"--from-backup", backupName, "--wait")
}

func (v *RestoreCmd) clusterRestoreDelete(restoreName string) []string {
	// if backupName is "*", delete all cluster restore
	if restoreName == "*" {
		return veleroCmdStr(Delete, Restore, all, confirm)
	}
	return veleroCmdStr(Delete, Restore, restoreName, confirm)
}

func (v *RestoreCmd) clusterRestoreList() []string {

	return veleroCmdStr(List, Restore, OutJson)
}

func (v *RestoreCmd) clusterRestoreDescribe(restoreName string) []string {
	// if backupName is "*", list describe all cluster restoreName
	if restoreName == "*" {
		return veleroCmdStr(Describe, Restore)
	}
	return veleroCmdStr(Describe, Restore, restoreName)
}
