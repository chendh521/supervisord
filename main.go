package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Configuration string `short:"c" long:"configuration" description:"the configuration file" default:"supervisord.conf"`
	Daemon        bool   `short:"d" long:"daemon" description:"run as daemon"`
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func initSignals(s *Supervisor) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{"signal": sig}).Info("receive a signal to stop all process & exit")
		s.procMgr.StopAllProcesses()
		os.Exit(-1)
	}()

}

var options Options
var parser = flags.NewParser(&options, flags.Default & ^flags.PrintErrors)

func RunServer() {
	// infinite loop for handling Restart ('reload' command)
	for true {
		s := NewSupervisor(options.Configuration)
		initSignals(s)
		if sErr := s.Reload(); sErr != nil {
			panic(sErr)
		}
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok {
			switch flagsErr.Type {
			case flags.ErrHelp:
				fmt.Fprintln(os.Stdout, err)
				os.Exit(0)
			case flags.ErrCommandRequired:
				if options.Daemon {
					Deamonize(RunServer)
				} else {
					RunServer()
				}
			default:
				panic(err)
			}
		}
	}
}
