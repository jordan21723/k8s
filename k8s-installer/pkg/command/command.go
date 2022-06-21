package command

import (
	"bytes"
	"fmt"
	"os/exec"

	"k8s-installer/pkg/log"
)

// run command
func RunCmd(command string, args ...string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(command, args...)
	var out, outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	log.Debug(fmt.Sprintf("Running command %s", cmd.String()))
	err := cmd.Run()
	if err != nil {
		return out, outErr, err
	}
	return out, outErr, nil
}

// run pipe command like cat a | grep -i "bla"
func CmdPipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}
	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}
	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}
