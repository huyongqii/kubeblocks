/*
Copyright ApeCloud, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"

	appsv1alpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	cfgutil "github.com/apecloud/kubeblocks/internal/configuration"
	cfgcontainer "github.com/apecloud/kubeblocks/internal/configuration/container"
)

var (
	logger = logr.Discard()
)

func SetLogger(zapLogger *zap.Logger) {
	logger = zapr.NewLogger(zapLogger)
	logger = logger.WithName("configmap_volume_watcher")
}

// findPidFromProcessName get parent pid
func findPidFromProcessName(processName string) (PID, error) {
	allProcess, err := process.Processes()
	if err != nil {
		return InvalidPID, err
	}

	psGraph := map[PID]int32{}
	for _, proc := range allProcess {
		name, err := proc.Name()
		// OS X getting the name of the system process sometimes fails,
		// because OS X Process.Name function depends on sysctl,
		// the function requires elevated permissions.
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed to get process name from pid[%d], and pass", proc.Pid))
			continue
		}
		if name != processName {
			continue
		}
		ppid, err := proc.Ppid()
		if err != nil {
			return InvalidPID, cfgutil.WrapError(err, "failed to get parent pid from pid[%d]", proc.Pid)
		}
		psGraph[PID(proc.Pid)] = ppid
	}

	for key, value := range psGraph {
		if _, ok := psGraph[PID(value)]; !ok {
			return key, nil
		}
	}

	return InvalidPID, cfgutil.MakeError("not find pid of process name: [%s]", processName)
}

func CreateSignalHandler(sig appsv1alpha1.SignalType, processName string) (WatchEventHandler, error) {
	signal, ok := allUnixSignals[sig]
	if !ok {
		err := cfgutil.MakeError("not support unix signal: %s", sig)
		logger.Error(err, "failed to create signal handler")
		return nil, err
	}
	return func(event fsnotify.Event) error {
		pid, err := findPidFromProcessName(processName)
		if err != nil {
			return err
		}
		logger.V(1).Info(fmt.Sprintf("find pid: %d from process name[%s]", pid, processName))
		return sendSignal(pid, signal)
	}, nil
}

func CreateExecHandler(command string) (WatchEventHandler, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		return nil, cfgutil.MakeError("invalid command: %s", command)
	}
	cmd := exec.Command(args[0], args[1:]...)
	return func(_ fsnotify.Event) error {
		stdout, err := cfgcontainer.ExecShellCommand(cmd)
		if err == nil {
			logger.V(1).Info(fmt.Sprintf("exec: [%s], result: [%s]", command, stdout))
		}
		return err
	}, nil
}

func IsValidUnixSignal(sig appsv1alpha1.SignalType) bool {
	_, ok := allUnixSignals[sig]
	return ok
}

func CreateTPLScriptHandler(tplScripts string, dirs []string, fileRegex string, backupPath string, formatConfig *appsv1alpha1.FormatterConfig) (WatchEventHandler, error) {
	logger.V(1).Info(fmt.Sprintf("config file regex: %s", fileRegex))
	logger.V(1).Info(fmt.Sprintf("config file reload script: %s", tplScripts))
	if _, err := os.Stat(tplScripts); err != nil {
		return nil, err
	}
	tplContent, err := os.ReadFile(tplScripts)
	if err != nil {
		return nil, err
	}
	if err := checkTPLScript(tplScripts, string(tplContent)); err != nil {
		return nil, err
	}
	filter, err := createFileRegex(fileRegex)
	if err != nil {
		return nil, err
	}
	if err := backupConfigFiles(dirs, filter, backupPath); err != nil {
		return nil, err
	}
	return func(event fsnotify.Event) error {
		var (
			lastVersion = []string{backupPath}
			currVersion = []string{filepath.Dir(event.Name)}
		)
		currFiles, err := scanConfigFiles(currVersion, filter)
		if err != nil {
			return err
		}
		lastFiles, err := scanConfigFiles(lastVersion, filter)
		if err != nil {
			return err
		}
		updatedParams, err := createUpdatedParamsPatch(currFiles, lastFiles, formatConfig)
		if err != nil {
			return err
		}
		if err := wrapGoTemplateRun(tplScripts, string(tplContent), updatedParams); err != nil {
			return err
		}
		return backupLastConfigFiles(currFiles, backupPath)
	}, nil
}
