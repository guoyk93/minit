package main

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// filename mark
// daily: FILENAME.ROT2020-06-02.EXT
// filesize: FILENAME.ROT000000000001.EXT (%012d)

const (
	LogrotateModeDaily    = "daily"
	LogrotateModeFilesize = "filesize"

	LogrotateDailyCron    = "5 0 * * *"
	LogrotateFilesizeCron = "@every 1m"

	LogrotateFilesize = 256 * 1024 * 1024

	RotationDateLayout = "2006-01-02"
)

var (
	rotationMarkPattern = regexp.MustCompile(`\.ROT(.+)\.`)
)

func rotationMarkRemove(filename string) string {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	base = rotationMarkPattern.ReplaceAllLiteralString(base, ".")
	return filepath.Join(dir, base)
}

func rotationMarkAdd(filename string, mark string) string {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	base = base[:len(base)-len(ext)] + ".ROT" + mark + ext
	return filepath.Join(dir, base)
}

type LogrotateRunner struct {
	files   []string
	mode    string
	keep    int
	dir     string
	command []string
	logger  *Logger
}

func (l *LogrotateRunner) Run(ctx context.Context) {
	l.logger.Printf("控制器启动")
	defer l.logger.Printf("控制器退出")

	var cronExpr string

	switch l.mode {
	case LogrotateModeDaily:
		cronExpr = LogrotateDailyCron
	case LogrotateModeFilesize:
		cronExpr = LogrotateFilesizeCron
	}

	cr := cron.New(cron.WithLogger(cron.PrintfLogger(l.logger)))
	_, err := cr.AddFunc(cronExpr, func() {
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
	now := time.Now()
	// 今天的零点，去掉时区信息
	bod := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// 昨天的零点，去掉时区信息
	boy := bod.Add(-time.Hour * 24)

	// 遍历所有通配符，建立原始文件列表，也就是移除 日期/分段 标记后的原始文件名
	baseFiles := map[string]bool{}
	for _, fPat := range l.files {
		files, _ := filepath.Glob(fPat)
		for _, file := range files {
			absFile, _ := filepath.Abs(file)
			if absFile != "" {
				baseFiles[rotationMarkRemove(absFile)] = true
			}
		}
	}

	// 遍历所有 baseFile，匹配所有已经轮转的文件
	for baseFile := range baseFiles {

		files, _ := filepath.Glob(rotationMarkAdd(baseFile, "*"))

		switch l.mode {
		case LogrotateModeDaily:
			// 按照从小到大排序，尽快暴露出日期最远的文件
			sort.Sort(sort.StringSlice(files))
		case LogrotateModeFilesize:
			// 按照从大到小排序，尽快暴露出编号最大的文件
			sort.Sort(sort.Reverse(sort.StringSlice(files)))
		}

		num := int64(0) // 最大编号
		rot := false    // 是否已经发现了昨天的日志文件

		// 遍历所有 baseFile 派生的文件
		for _, file := range files {

			// 寻找 ROT 标记
			var mark string
			if subs := rotationMarkPattern.FindStringSubmatch(filepath.Base(file)); len(subs) != 2 {
				continue
			} else {
				mark = subs[1]
			}

			switch l.mode {
			case LogrotateModeDaily:
				// 日期模式
				var date time.Time
				var err error
				if date, err = time.Parse(RotationDateLayout, mark); err != nil {
					l.logger.Errorf("无法进行日期匹配: %s: %s", file, err.Error())
					continue
				}
				if bod.Sub(date) > time.Hour*24*time.Duration(l.keep) {
					// 寻找过久的文件，进行删除
					if err := os.Remove(file); err != nil {
						l.logger.Errorf("删除文件失败: %s: %s", file, err.Error())
					} else {
						l.logger.Printf("删除文件: %s", file)
					}
				} else if boy.Sub(date) == 0 {
					// 如果昨日文件已经生成，则设置标记位
					rot = true
					l.logger.Printf("昨日文件已经生成：%s", file)
				} else {
					l.logger.Printf("忽略文件: %s", file)
				}
			case LogrotateModeFilesize:
				// 文件大小截断模式
				var id int64
				var err error
				if id, err = strconv.ParseInt(strings.TrimPrefix(mark, "0"), 10, 64); err != nil {
					l.logger.Errorf("无法进行编号匹配: %s: %s", file, err.Error())
					continue
				}
				if id > num {
					// 寻找最大编号，因为已经排序，因此第一个编号即是最大编号
					num = id
				} else if num-id > int64(l.keep) {
					// 删除过旧的文件
					if err := os.Remove(file); err != nil {
						l.logger.Errorf("删除文件失败: %s: %s", file, err.Error())
					} else {
						l.logger.Printf("删除文件: %s", file)
					}
				} else {
					l.logger.Printf("忽略文件: %s", file)
				}
			}

			// 完成对 baseFile 派生文件的处理
		}

		// 对 baseFile 进行 rotation
		switch l.mode {
		case LogrotateModeDaily:
			// 如果已经出现昨天的日志文件，则不 rotation
			if !rot {
				nname := rotationMarkAdd(baseFile, boy.Format(RotationDateLayout))
				if err := os.Rename(baseFile, nname); err != nil {
					l.logger.Errorf("无法重命名文件: %s -> %s: %s", baseFile, nname, err.Error())
					continue
				} else {
					l.logger.Printf("已经重命名文件: %s -> %s", baseFile, nname)
				}
			}
		case LogrotateModeFilesize:
			fi, err := os.Stat(baseFile)
			if err != nil {
				l.logger.Errorf("无法检查文件大小: %s: %s", baseFile, err.Error())
				continue
			}
			// 如果 baseFile 文件过大，则进行 rotation
			if fi.Size() > LogrotateFilesize {
				nname := rotationMarkAdd(baseFile, fmt.Sprintf("%012d", num+1))
				if err := os.Rename(baseFile, nname); err != nil {
					l.logger.Errorf("无法重命名文件: %s -> %s: %s", baseFile, nname, err.Error())
					continue
				} else {
					l.logger.Printf("已经重命名文件: %s -> %s", baseFile, nname)
				}
			}
		}

		// 执行后续命令
		if err := execute(l.dir, l.command, l.logger); err != nil {
			l.logger.Errorf("命令启动失败: %s", err.Error())
			return
		}
	}
}

func NewLogrotateRunner(files []string, mode string, keep int, dir string, command []string, logger *Logger) (Runner, error) {
	switch mode {
	case LogrotateModeDaily:
	case LogrotateModeFilesize:
	default:
		return nil, fmt.Errorf("未知的 logrotate 模式: %s", mode)
	}
	return &LogrotateRunner{
		files:   files,
		mode:    mode,
		keep:    keep,
		dir:     dir,
		command: command,
		logger:  logger,
	}, nil
}
