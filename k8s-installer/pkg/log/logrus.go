package log

import (
	"io"
	"net"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var nodeIp net.IP

func DefaultLoginSetting(getNodeIp func(bool) net.IP, logRetentionPeriod int) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.TraceLevel)

	nodeIp = getNodeIp(true)

	writer, err := rotatelogs.New(
		"/var/log/caas4.log",
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(logRetentionPeriod)),
		rotatelogs.WithRotationTime(time.Hour*24*time.Duration(logRetentionPeriod)))

	if err != nil {
		logrus.Fatalf("error opening file: %v", err)
	}
	mw := io.MultiWriter(writer, os.Stdout)
	logrus.SetOutput(mw)
}

func SetLogLevel(level uint32) {
	logrus.SetLevel(logrus.Level(level))
}

func Debug(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Debugf(format, args...)
}

func Info(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Infof(format, args...)
}

func Fatal(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Fatalf(format, args...)
}

func Error(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Errorf(format, args...)
}

func Warn(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		"ip": nodeIp,
	}).Warnf(format, args...)
}
