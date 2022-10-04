// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestTerminateSync(t *testing.T) {
	cmd := exec.Command("/bin/sleep", "20")
	require.NoError(t, cmd.Start())
	require.NotNil(t, cmd.Process)
	err := TerminateSync(cmd.Process.Pid, time.After(10*time.Second))
	require.NoError(t, err)
	require.Equal(t, os.ErrProcessDone, cmd.Process.Signal(unix.SIGHUP))
}

func TestTerminateSync_ignoring_process(t *testing.T) {

	tests := []struct {
		processTimeSeconds int
		gracePeriod        time.Duration
		fileExists         bool
	}{
		{
			processTimeSeconds: 1,
			gracePeriod:        7 * time.Second,
			fileExists:         true,
		},
		{
			processTimeSeconds: 7,
			gracePeriod:        time.Second,
			fileExists:         false,
		},
	}
	for _, test := range tests {
		dir, err := ioutil.TempDir("", "terminal_test_close")
		require.NoError(t, err)
		expectedFile := dir + "/done.txt"
		script := dir + "/script.sh"
		err = ioutil.WriteFile(script, []byte(fmt.Sprintf(`#!/bin/bash
		trap 'echo \"Be patient\"' SIGTERM SIGINT SIGHUP
		for ((n= %d ; n; n--))
		do
		   sleep 1
		done
		echo touching
		touch %s
		echo touched
		`, test.processTimeSeconds, expectedFile)), 0644)
		require.NoError(t, err)

		cmd := exec.Command("/bin/bash", script)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Start())
		require.NotNil(t, cmd.Process)
		time.Sleep(100 * time.Millisecond)
		require.NotEqual(t, os.ErrProcessDone, cmd.Process.Signal(syscall.Signal(0)))
		err = TerminateSync(cmd.Process.Pid, time.After(test.gracePeriod))
		if test.fileExists {
			require.NoError(t, err)
			require.FileExists(t, expectedFile)
		} else {
			require.Equal(t, ErrForceKilled, err)
			require.NoFileExists(t, expectedFile)
		}
	}
}
