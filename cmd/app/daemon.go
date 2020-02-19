package app

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"storage-manager/pkg/api"
	//"storage-manager/pkg/k8s"
)

func DaemonCmd() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			if err := startManager(c); err != nil {
				logrus.Fatalf("Failed to start manager daemon, err: %v", err)
			}
			return nil
		},
	}
}

func startManager(c *cli.Context) error {

	//client := k8s.NewK8SClient("", "")
	//logrus.Infof("zzlin test client: %v", client)

	go api.StartHttpServer()

	return nil
}
