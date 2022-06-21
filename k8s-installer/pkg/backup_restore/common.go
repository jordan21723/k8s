package backup_restore

import (
	"fmt"
	"k8s-installer/pkg/log"
	"strings"
	"time"
)

const (
	velero = "velero"

	Backup        = "backup"
	BackupRegular = "schedule"
	Restore       = "restore"

	Create   = "create"
	Delete   = "delete"
	List     = "get"
	Describe = "describe"

	Cluster   = "cluster"
	Namespace = "cluster"

	all     = "--all"
	confirm = "--confirm"

	OutTable = "-o table"
	OutJson  = "-o json"
	OutYaml  = "-o yaml"
)

// TODO: namespace backup and restore

// TODO: other backup and restore

func veleroCmdStr(action, available string, arg ...string) []string {
	c := fmt.Sprintf("%s %s %s %s", velero, action, available, strings.Join(arg, " "))

	log.Info(c)
	return strings.Fields(c)
}

func setName(available string) string {
	return strings.ReplaceAll(available+time.Now().Format("2006-01-02-15-04-05"), "-", "")
}
