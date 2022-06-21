package sshutils

import (
	"fmt"
	"k8s-installer/pkg/util/rsautils"
	"k8s-installer/schema"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHClient(credential *schema.SSHCredential) (*ssh.Client, error) {
	username, err := rsautils.RsaDecrypt(credential.Username)
	if err != nil {
		return nil, err
	}
	password, err := rsautils.RsaDecrypt(credential.Password)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		Timeout:         5 * time.Second,
		User:            string(username),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(string(password))},
	}
	addr := fmt.Sprintf("%s:%d", credential.IpAddress, credential.Port)
	c, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}
