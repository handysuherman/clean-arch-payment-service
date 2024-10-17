package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/infra"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type flags struct {
	appUniqueID                 string
	appEnvironment              string
	appHostname                 string
	appPublishedPort            int
	appHealthCheckPublishedAddr string
	appHealthCheckPublishedPort int

	configFilePath string
	etcdTimeout    time.Duration
}

func main() {
	cmd, err := exportedFlags()
	if err != nil {
		fmt.Println(err)
		return
	}
	logs := logger.NewLogger()

	if cmd.appEnvironment == "develop" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cm := config.NewManager(logs, cmd.etcdTimeout)
	cm.SetConsulAppSrvConfig(cmd.appUniqueID, cmd.appPublishedPort, cmd.appHostname, cmd.appHealthCheckPublishedAddr, cmd.appHealthCheckPublishedPort)

	cfg, err := cm.Bootstrap(cmd.configFilePath)
	if err != nil {
		logs.Debug(err)
		return
	}

	_app := infra.New(logs, cfg, cm)
	logs.Fatal(_app.Run())
}

func exportedFlags() (*flags, error) {
	appConfigFilePath := flag.String("config-file", "./config.yaml", "app configuration file path")
	etcdTimeout := flag.String("etcd-timeout", "15s", "etcd dial timeout")

	appEnvironment := flag.String("env", "develop", "app environment")
	appHostname := flag.String("hostname", "", "app hostname")
	appUniqueID := flag.String("uid", "", "app unique id")
	appPublishedPort := flag.Int("published-port", 0, "app published port")
	appHealthCheckPublishedAddr := flag.String("health-check-addr", "", "app healthCheck published addr")
	appHealthCheckPublishedPort := flag.Int("health-check-published-port", 0, "app healthCheck published port")

	flag.Parse()

	if *appUniqueID == "" {
		return nil, fmt.Errorf("uid flag is required")
	}

	if *appHostname == "" {
		return nil, fmt.Errorf("hostname flag is required")
	}

	if *appPublishedPort == 0 {
		return nil, fmt.Errorf("published-port flag is required")
	}

	if *appHealthCheckPublishedAddr == "" {
		return nil, fmt.Errorf("health-check-addr flag is required")
	}

	if *appHealthCheckPublishedPort == 0 {
		return nil, fmt.Errorf("health-check-published-port flag is required")
	}

	timeoutDur, err := time.ParseDuration(*etcdTimeout)
	if err != nil {
		return nil, fmt.Errorf("error occurred when parsing etcd dial timeout: %v", err)
	}

	return &flags{
		appUniqueID:                 *appUniqueID,
		appEnvironment:              *appEnvironment,
		appHostname:                 *appHostname,
		appPublishedPort:            *appPublishedPort,
		appHealthCheckPublishedAddr: *appHealthCheckPublishedAddr,
		appHealthCheckPublishedPort: *appHealthCheckPublishedPort,
		configFilePath:              *appConfigFilePath,
		etcdTimeout:                 timeoutDur,
	}, nil
}
