package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/guoyk93/minit/pkg/mlog"
	"github.com/guoyk93/minit/pkg/tmplfuncs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type RenderRunner struct {
	Unit
	logger *mlog.Logger
}

func (r *RenderRunner) Run(ctx context.Context) {
	r.logger.Printf("控制器启动")
	defer r.logger.Printf("控制器退出")

	env := environ()

	for _, filePattern := range r.Files {
		var err error
		var names []string
		if names, err = filepath.Glob(filePattern); err != nil {
			r.logger.Errorf("匹配表达式 %s 格式错误: %s", filePattern, err.Error())
			continue
		}
		for _, name := range names {
			var buf []byte
			if buf, err = ioutil.ReadFile(name); err != nil {
				r.logger.Errorf("无法读取文件: %s", name)
				continue
			}
			tmpl := template.New("__main__").Funcs(tmplfuncs.Funcs).Option("missingkey=zero")
			if tmpl, err = tmpl.Parse(string(buf)); err != nil {
				r.logger.Errorf("无法解析文件 %s: %s", name, err.Error())
				continue
			}
			out := &bytes.Buffer{}
			if err = tmpl.Execute(out, map[string]interface{}{
				"Env": env,
			}); err != nil {
				r.logger.Errorf("无法渲染文件 %s: %s", name, err.Error())
				continue
			}
			if err = ioutil.WriteFile(name, out.Bytes(), 0755); err != nil {
				r.logger.Errorf("无法写入文件 %s: %s", name, err.Error())
				continue
			}
			r.logger.Printf("文件渲染完成: %s", name)
		}
	}
}

func NewRenderRunner(unit Unit, logger *mlog.Logger) (Runner, error) {
	if len(unit.Files) == 0 {
		return nil, fmt.Errorf("没有指定文件，检查 files 字段")
	}
	return &RenderRunner{
		Unit:   unit,
		logger: logger,
	}, nil
}

func environ() map[string]string {
	out := make(map[string]string)
	envs := os.Environ()
	for _, entry := range envs {
		splits := strings.SplitN(entry, "=", 2)
		if len(splits) == 2 {
			out[strings.TrimSpace(splits[0])] = strings.TrimSpace(splits[1])
		}
	}
	return out
}
