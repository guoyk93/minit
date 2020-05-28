package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type RunnerLevel int

const (
	RunnerL1 RunnerLevel = iota + 1
	RunnerL2
	RunnerL3
)

type RunnerFactory struct {
	Level  RunnerLevel
	Create func(unit Unit, logger *Logger) (Runner, error)
}

var (
	RunnerFactories = map[string]*RunnerFactory{
		"render": {
			Level: RunnerL1,
			Create: func(unit Unit, logger *Logger) (Runner, error) {
				return NewRenderRunner(unit.Files, logger)
			},
		},
		"once": {
			Level: RunnerL2,
			Create: func(unit Unit, logger *Logger) (Runner, error) {
				return NewOnceRunner(unit.Dir, unit.Command, logger)
			},
		},
		"daemon": {
			Level: RunnerL3,
			Create: func(unit Unit, logger *Logger) (Runner, error) {
				return NewDaemonRunner(unit.Dir, unit.Command, logger)
			},
		},
		"cron": {
			Level: RunnerL3,
			Create: func(unit Unit, logger *Logger) (Runner, error) {
				return NewCronRunner(unit.Cron, unit.Dir, unit.Command, logger)
			},
		},
	}
)

type Runner interface {
	Run(ctx context.Context)
}

var (
	childPids                 = map[int]bool{}
	childPidsLock sync.Locker = &sync.Mutex{}
)

func addPid(pid int) {
	childPidsLock.Lock()
	defer childPidsLock.Unlock()
	childPids[pid] = true
}

func removePid(pid int) {
	childPidsLock.Lock()
	defer childPidsLock.Unlock()
	delete(childPids, pid)
}

func notifyPIDs(sig os.Signal) {
	childPidsLock.Lock()
	defer childPidsLock.Unlock()
	for pid, found := range childPids {
		if found {
			if process, _ := os.FindProcess(pid); process != nil {
				_ = process.Signal(sig)
			}
		}
	}
}

func execute(dir string, command []string, logger *Logger) (err error) {
	// 为命令行注入环境变量
	argv := make([]string, 0, len(command))
	for _, arg := range command {
		argv = append(argv, os.ExpandEnv(arg))
	}
	// 构建 cmd
	var outPipe, errPipe io.ReadCloser
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	// 阻止信号传递
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if outPipe, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if errPipe, err = cmd.StderrPipe(); err != nil {
		return
	}

	// 执行
	if err = cmd.Start(); err != nil {
		return
	}

	// 记录 Pid
	addPid(cmd.Process.Pid)

	// 串流
	go logger.StreamOut(outPipe)
	go logger.StreamErr(errPipe)

	// 等待退出
	if err = cmd.Wait(); err != nil {
		logger.Errorf("进程退出: %s", err.Error())
		err = nil
	} else {
		logger.Printf("进程退出")
	}

	// 移除 Pid
	removePid(cmd.Process.Pid)

	return
}
