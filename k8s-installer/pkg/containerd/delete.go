package containerd

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"k8s-installer/pkg/log"
	"syscall"
)

func DeleteAllContainer(namespace string) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), namespace)

	if ctrs, err := client.Containers(ctx); err != nil {
		return err
	} else {
		for _, ctr := range ctrs {
			log.Debugf("Attempt to kill and delete task belong to container %s", ctr.ID())
			if task, err := ctr.Task(ctx, nil); err != nil {
				log.Warnf("Failed to get task of container %s , it may has not task at all, let move on", ctr.ID())
			} else {
				log.Debugf("Attempt to kill task %s", task.ID())
				exitStatusC, err := task.Wait(ctx)
				if err != nil {
					log.Errorf("Failed to get status chan due to error %s", err)
					return err
				}
				if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
					log.Errorf("Failed to kill task due to error %s", err)
					return err
				}
				log.Debugf("Containerd task %s killed", task.ID())
				log.Debugf("Wait for task %s exit signal", task.ID())
				status := <-exitStatusC
				log.Debugf("Got signal from task %s", task.ID())
				code, _, err := status.Result()
				if err != nil {
					log.Errorf("Failed to get task result due to error %s", err)
					return err
				}
				log.Debugf("Got task exit signal %v", code)
				log.Debugf("Attempt to delete task %s", task.ID())
				if statusCode, err := task.Delete(ctx); err != nil {
					log.Errorf("(ignore) Failed to delete task %s due to error: %s since task already been killed it`s ok to leave it alone", task.ID(), err)
				} else {
					log.Debugf("Task %s deleted and got exit code %v", task.ID(), statusCode.ExitCode())
				}
			}
			log.Debugf("Attempt to delete container %s", ctr.ID())
			if err := ctr.Delete(ctx); err != nil {
				log.Errorf("(ignored) Failed to delete container %s due to error: %s but container without task can be consider harmless let`s move on", ctr.ID(), err.Error())
				continue
			}
		}
	}
	return nil
}
