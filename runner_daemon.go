package main

import (
	"context"
	"fmt"
	"time"
)

type DaemonRunner struct {
	dir     string
	command []string
	logger  *Logger
}

func (r *DaemonRunner) Run(ctx context.Context) {
	r.logger.Printf("控制器启动")
	defer r.logger.Printf("控制器退出")
forLoop:
	for {
		// 检查 ctx 是否已经结束
		if ctx.Err() != nil {
			break forLoop
		}

		var err error
		if err = execute(r.dir, r.command, r.logger); err != nil {
			r.logger.Errorf("启动失败: %s", err.Error())
		}

		// 检查 ctx 是否已经结束
		if ctx.Err() != nil {
			break forLoop
		}

		// 重试
		r.logger.Printf("5s 后重启")

		timer := time.NewTimer(time.Second * 5)
		select {
		case <-timer.C:
		case <-ctx.Done():
			break forLoop
		}
	}
}

func NewDaemonRunner(dir string, command []string, logger *Logger) (Runner, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("没有指定命令，检查 command 字段")
	}
	return &DaemonRunner{
		dir:     dir,
		command: command,
		logger:  logger,
	}, nil
}