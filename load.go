package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FilterMode int

const (
	FilterNone FilterMode = iota
	FilterWhitelist
	FilterBlacklist
)

var (
	filters    = map[string]bool{}
	filterMode FilterMode
)

func init() {
	filterMode = FilterNone

	var raw string
	if raw = os.Getenv("MINIT_ENABLE"); raw != "" {
		filterMode = FilterWhitelist
	} else if raw = os.Getenv("MINIT_DISABLE"); raw != "" {
		filterMode = FilterBlacklist
	}

	names := strings.Split(raw, ",")
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "@" {
			continue
		}
		filters[name] = true
	}
}

type Unit struct {
	Name    string   `yaml:"name"`
	Group   string   `yaml:"group"`
	Kind    string   `yaml:"kind"`
	Cron    string   `yaml:"cron"`
	Dir     string   `yaml:"dir"`
	Files   []string `yaml:"files"`
	Count   int      `yaml:"count"`
	Command []string `yaml:"command"`
}

func (u Unit) CanonicalName() string {
	return u.Kind + "/" + u.Name
}

func LoadDir(dir string) (units []Unit, err error) {
	var files []string
	for _, ext := range []string{"*.yml", "*.yaml"} {
		if files, err = filepath.Glob(filepath.Join(dir, ext)); err != nil {
			return
		}
		for _, file := range files {
			var units0 []Unit
			if units0, err = LoadFile(file); err != nil {
				return
			}
			units = append(units, units0...)
		}
	}
	return
}

func LoadFile(fn string) (units []Unit, err error) {
	var f *os.File
	if f, err = os.Open(fn); err != nil {
		return
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	for {
		var unit Unit
		if err = dec.Decode(&unit); err != nil {
			if err == io.EOF {
				err = nil
			} else {
				err = fmt.Errorf("无法解析文件 %s: %s", fn, err.Error())
			}
			return
		}

		// 清理下空格
		unit.Name = strings.TrimSpace(unit.Name)
		unit.Kind = strings.TrimSpace(unit.Kind)
		unit.Cron = strings.TrimSpace(unit.Cron)
		unit.Dir = strings.TrimSpace(unit.Dir)
		unit.Group = strings.TrimSpace(unit.Group)

		// 打开关闭
		switch filterMode {
		case FilterNone:
		case FilterWhitelist:
			if !filters[unit.Name] && !filters["@"+unit.Group] {
				log.Printf("取消单元载入: %s", unit.Name)
				continue
			}
		case FilterBlacklist:
			if filters[unit.Name] || filters["@"+unit.Group] {
				log.Printf("取消单元载入: %s", unit.Name)
				continue
			}
		}

		// 重复型
		if unit.Count > 0 {
			for i := 0; i < unit.Count; i++ {
				subUnit := unit
				subUnit.Name = fmt.Sprintf("%s-%d", unit.Name, i+1)
				units = append(units, subUnit)
			}
		} else {
			units = append(units, unit)
		}
	}
}
