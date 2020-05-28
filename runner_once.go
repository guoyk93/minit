package main

import (
	"context"
	"fmt"
)

type OnceRunner struct {
	dir     string
	command []string
	logger  *Logger
}

func (r *OnceRunner) Run(ctx context.Context) {
	r.logger.Printf("控制器启动")
	defer r.logger.Printf("控制器退出")
	if err := execute(r.dir, r.command, r.logger); err != nil {
		r.logger.Errorf("启动失败: %s", err.Error())
		return
	}
}

func NewOnceRunner(dir string, command []string, logger *Logger) (Runner, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("没有指定命令，检查 command 字段")
	}
	return &OnceRunner{
		dir:     dir,
		command: command,
		logger:  logger,
	}, nil
}
