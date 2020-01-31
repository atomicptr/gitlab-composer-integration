package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"

	"github.com/atomicptr/gitlab-composer-integration/service"
)

const ConfNamespace = ""

func main() {
	err := run()
	if err != nil {
		log.Printf("error: %s", err)
	}
}

func run() error {
	// logger
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// configuration
	var config service.Config

	if err := conf.Parse(os.Args[1:], ConfNamespace, &config); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage(ConfNamespace, &config)
			if err != nil {
				return errors.Wrap(err, "generating usage")
			}

			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "error: parsing config")
	}

	// channel to listen for errors coming from the service
	serviceErrors := make(chan error, 1)

	// service starting
	logger.Printf("main: gitlab-composer-integration starting...")
	defer logger.Printf("main: Done")

	out, err := conf.String(&config)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	logger.Printf("main: Config:\n%v\n", out)

	// channel to listen for interrupt or terminate signal from OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	svc := service.New(
		service.Config{
			Port: config.Port,
		},
		logger,
		serviceErrors,
	)

	go func() {
		serviceErrors <- svc.Run()
	}()

	select {
	case err := <-serviceErrors:
		return errors.Wrap(err, "service error")
	case sig := <-shutdown:
		logger.Printf("main: %v shutdown...", sig)

		err := svc.Stop()
		if err != nil {
			logger.Printf("error: %s", err)
		}

		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		}
	}

	return nil

}
