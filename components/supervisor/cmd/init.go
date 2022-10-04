// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/gitpod-io/gitpod/common-go/process"
	"github.com/gitpod-io/gitpod/supervisor/pkg/supervisor"
	"github.com/prometheus/procfs"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init the supervisor",
	Run: func(cmd *cobra.Command, args []string) {
		log.Init(ServiceName, Version, true, false)
		cfg, err := supervisor.GetConfig()
		if err != nil {
			log.WithError(err).Info("cannnot load config")
		}
		var (
			sigInput = make(chan os.Signal, 1)
		)
		signal.Notify(sigInput, os.Interrupt, syscall.SIGTERM)

		supervisorPath, err := os.Executable()
		if err != nil {
			supervisorPath = "/.supervisor/supervisor"
		}
		runCommand := exec.Command(supervisorPath, "run")
		runCommand.Args[0] = "supervisor"
		runCommand.Stdin = os.Stdin
		runCommand.Stdout = os.Stdout
		runCommand.Stderr = os.Stderr
		runCommand.Env = os.Environ()
		err = runCommand.Start()
		if err != nil {
			log.WithError(err).Error("supervisor run start error")
			return
		}

		supervisorDone := make(chan struct{})
		go func() {
			defer close(supervisorDone)

			err := runCommand.Wait()
			if err != nil && !(strings.Contains(err.Error(), "signal: interrupt") || strings.Contains(err.Error(), "no child processes")) {
				log.WithError(err).Error("supervisor run error")
				return
			}
		}()

		select {
		case <-supervisorDone:
			// supervisor has ended - we're all done here
			return
		case <-sigInput:
			// we received a terminating signal - pass on to supervisor and wait for it to finish
			timer := time.After(cfg.GetTerminationGracePeriod())
			terminationDone := make(chan int)
			go func() {
				defer close(terminationDone)
				err := process.TerminateSync(runCommand.Process.Pid, timer)
				if err != nil {
					log.WithError(err).Error("supervisor termination error")
				}

				// now terminate reparented processes until there are none anymore or the time is up
				for {
					processes, err := procfs.AllProcs()
					if err != nil {
						log.WithError(err).Error("Cannot obtain processes")
					}
					if len(processes) == 1 {
						close(supervisorDone)
						return
					}
					// send SIGTERM to all processes but ourself
					g := new(errgroup.Group)
					for _, proc := range processes {
						if proc.PID == os.Getpid() {
							continue
						}
						p := proc
						g.Go(func() error {
							return process.TerminateSync(p.PID, timer)
						})
					}
					err = g.Wait()
					if err != nil {
						log.WithError(err).Error("terminating child processes")
					}
				}

			}()
			// wait for either successful termination or the timeout
			select {
			case <-timer:
				// Time is up, but we give all the goroutines a bit more time to react to this.
				time.Sleep(time.Millisecond * 500)
			case <-terminationDone:
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
