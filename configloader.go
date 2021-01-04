package main

import (
	"io/ioutil"

	"github.com/goccy/go-yaml"
)

type YamlConfLoader struct {
	Targets      []string `yaml: "targets"`
	BackupDir    string   `yaml: "backupdir"`
	HistoryCount int      `yaml: "historycount"`
	TargetSuffix []string `yaml: "targetsuffix"`
}

func (cl *YamlConfLoader) Load(fname string) error {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, cl)
	if err != nil {
		return err
	}
	return nil
}

func (cl *YamlConfLoader) GetTgts() []string {
	return cl.Targets
}

func (cl *YamlConfLoader) GetBackupDir() string {
	return cl.BackupDir
}

func (cl *YamlConfLoader) GetHistoryCount() int {
	return cl.HistoryCount
}

func (cl *YamlConfLoader) GetTgtSuffix() []string {
	return cl.TargetSuffix
}
