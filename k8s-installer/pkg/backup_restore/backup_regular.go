package backup_restore

type BackupRegularCmd struct {
	Action            string
	CronJobTime       string
	Output            string
	BackupRegularName string
	Args              []string
}

func NewBackupRegularCmd(action, backupRegularName, cronJobTime, output string, args []string) *BackupRegularCmd {
	return &BackupRegularCmd{
		Action:            action,
		BackupRegularName: backupRegularName,
		CronJobTime:       cronJobTime,
		Output:            output,
		Args:              args,
	}
}

func (v *BackupRegularCmd) ActionCmd() []string {
	switch v.Action {
	case Create:
		return v.clusterBackupRegularCreate()
	case Delete:
		return v.clusterBackupRegularDelete(v.BackupRegularName)
	case List:
		return v.clusterBackupRegularList(v.Output)
	case Describe:
		return v.clusterBackupRegularDescribe(v.BackupRegularName)
	default:
		return []string{}
	}
}

func (v *BackupRegularCmd) clusterBackupRegularCreate() []string {
	if v.BackupRegularName == "" {
		v.BackupRegularName = setName(Backup)
	}

	return veleroCmdStr(Create, BackupRegular, v.BackupRegularName, "--schedule="+v.CronJobTime,
		"--include-cluster-resources", "--exclude-namespaces", "minio,velero")
}

func (v *BackupRegularCmd) clusterBackupRegularDelete(backupRegularName string) []string {
	// if backupName is "*", delete all cluster backup regular
	if backupRegularName == "*" {
		return veleroCmdStr(Delete, BackupRegular, all, confirm)
	}
	return veleroCmdStr(Delete, BackupRegular, backupRegularName, confirm)
}

func (v *BackupRegularCmd) clusterBackupRegularList(output string) []string {
	if output == "" {
		output = OutJson
	}
	return veleroCmdStr(List, BackupRegular, OutJson)
}

func (v *BackupRegularCmd) clusterBackupRegularDescribe(backupRegularName string) []string {
	// if backupName is "*", list describe all cluster backup regular
	if backupRegularName == "*" {
		return veleroCmdStr(Describe, BackupRegular)
	}
	return veleroCmdStr(Describe, BackupRegular, backupRegularName)
}
