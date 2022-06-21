package backup_restore

type BackupCmd struct {
	Action     string
	BackupName string
	Output     string
	Args       []string
}

func NewBackupCmd(action, backupName, output string, args []string) *BackupCmd {
	return &BackupCmd{
		Action:     action,
		BackupName: backupName,
		Output:     output,
		Args:       args,
	}
}

func (v *BackupCmd) ActionCmd() []string {
	switch v.Action {
	case Create:
		return v.clusterBackupCreate()
	case Delete:
		return v.clusterBackupDelete(v.BackupName)
	case List:
		return v.clusterBackupList(v.Output)
	case Describe:
		return v.clusterBackupDescribe(v.BackupName)
	default:
		return []string{}
	}
}

func (v *BackupCmd) clusterBackupCreate() []string {
	if v.BackupName == "" {
		v.BackupName = setName(Backup)
	}
	return veleroCmdStr(Create, Backup, v.BackupName,
		"--include-cluster-resources", "--exclude-namespaces", "minio,velero", "--wait")
}

func (v *BackupCmd) clusterBackupDelete(backupName string) []string {
	// if backupName is "*", delete all cluster backup
	if backupName == "*" {
		return veleroCmdStr(Delete, Backup, all, confirm)
	}
	return veleroCmdStr(Delete, Backup, backupName, confirm)
}

func (v *BackupCmd) clusterBackupList(output string) []string {
	if output == "" {
		output = OutJson
	}
	return veleroCmdStr(List, Backup, output)
}

func (v *BackupCmd) clusterBackupDescribe(backupName string) []string {
	// if backupName is "*", list describe all cluster backup
	if backupName == "*" {
		return veleroCmdStr(Describe, Backup)
	}
	return veleroCmdStr(Describe, Backup, backupName)
}
