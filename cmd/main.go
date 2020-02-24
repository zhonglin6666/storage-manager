package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"storage-manager/cmd/app"
)

var (
	version = "v0.0.1"
)

const (
	helpUsage = `

	`
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	app := cli.App{
		Name:            "manager",
		Usage:           helpUsage,
		Version:         version,
		CommandNotFound: cmdNotFound,
		Before: func(ctx *cli.Context) error {
			if ctx.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug, d",
				Usage: "enable debug log level",
			},
			&cli.BoolFlag{
				Name:  "memory, m",
				Usage: "implement file system with memory",
			},
		},
		Commands: []*cli.Command{
			app.DaemonCmd(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf("Run error: %v", err)
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logrus.Infof("Shutdown signal, shutting down http service ...")
}

func cmdNotFound(ctx *cli.Context, cmd string) {
	panic(fmt.Errorf("invalid command: %s", cmd))
}
