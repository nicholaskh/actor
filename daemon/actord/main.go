package main

import (
	"fmt"
	"github.com/funkygao/actor/actor"
	"github.com/funkygao/golib/locking"
	"github.com/funkygao/golib/server"
	"github.com/funkygao/golib/signal"
	log "github.com/funkygao/log4go"
	"math/rand"
	"os"
	"runtime/debug"
	"syscall"
	"time"
)

var (
	actorRunner *actor.Actor
)

func init() {
	parseFlags()

	if options.showVersion {
		server.ShowVersionAndExit()
	}

	if options.kill {
		s := server.NewServer("actord")
		s.LoadConfig(options.configFile)
		s.Launch()

		actor.New(s).Stop()

		if err := server.KillProcess(options.lockFile); err != nil {
			fmt.Fprintf(os.Stderr, "stop failed: %s\n", err)
		}

		os.Exit(0)
	}

	server.SetupLogging(options.logFile, options.logLevel, options.crashLogFile)

	if options.lockFile != "" {
		if locking.InstanceLocked(options.lockFile) {
			fmt.Fprintf(os.Stderr, "Another actor is running, exit...\n")
			os.Exit(1)
		}

		locking.LockInstance(options.lockFile)
	}

	signal.RegisterSignalHandler(syscall.SIGINT, func(sig os.Signal) {
		shutdown()
	})

	rand.Seed(time.Now().UnixNano())
}

func main() {
	defer func() {
		cleanup()

		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()

		}
	}()

	log.Info("%s", `
                _             
               | |            
      __ _  ___| |_ ___  _ __ 
     / _  |/ __| __/ _ \| '__|
    | (_| | (__| || (_) | |   
     \__,_|\___|\__\___/|_| `)

	s := server.NewServer("actord")
	s.LoadConfig(options.configFile)
	s.Launch()

	actorRunner = actor.New(s)
	actorRunner.ServeForever()

	shutdown()
}
