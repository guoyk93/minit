package main

import (
	"context"
	"github.com/robfig/cron/v3"
)

const (
	LogrotateCron = "5 0 * * *"
)

type LogrotateRunner struct {
	files   []string
	dir     string
	keep    int
	command []string
	logger  *Logger
}

func (l *LogrotateRunner) Run(ctx context.Context) {
	l.logger.Printf("控制器启动")
	defer l.logger.Printf("控制器退出")

	cr := cron.New(cron.WithLogger(cron.PrintfLogger(l.logger)))
	_, err := cr.AddFunc(LogrotateCron, func() {
		l.logger.Printf("开始日志轮转")
		defer l.logger.Printf("结束日志轮转")
		l.rotate()
	})
	if err != nil {
		panic(err)
	}

	cr.Start()
	<-ctx.Done()
	<-cr.Stop().Done()
}

func (l *LogrotateRunner) rotate() {
	// TODO: implements
}

func NewLogrotateRunner(files []string, keep int, dir string, command []string, logger *Logger) (Runner, error) {
	return &LogrotateRunner{
		files:   files,
		keep:    keep,
		dir:     dir,
		command: command,
		logger:  logger,
	}, nil
}
