package app

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"storage-manager/pkg/api"
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

	api.StartHttpServer()

	return nil
}
