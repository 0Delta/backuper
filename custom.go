package main

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type EmyaHandler struct {
	BackupDir    string
	HistoryCount int
	TgtSuffix    []string
}

func (h *EmyaHandler) Init(conf HandlerConfigLoader) error {
	h.BackupDir = conf.GetBackupDir()
	log.Println("BackupDir: ", h.BackupDir)
	if err := os.MkdirAll(filepath.Dir(h.BackupDir), 0777); err != nil {
		return err
	}
	h.HistoryCount = conf.GetHistoryCount()
	log.Println("HistoryCount: ", h.HistoryCount)

	h.TgtSuffix = conf.GetTgtSuffix()
	if 1 > len(h.TgtSuffix) {
		return errors.New("NoTarget")
	}
	log.Println("Target: ", h.TgtSuffix)

	return nil
}

func (h *EmyaHandler) CheckTarget(fname string) bool {
	flg := false
	for _, suf := range h.TgtSuffix {
		flg = flg || strings.HasSuffix(fname, suf)
	}
	return flg
}

func (h *EmyaHandler) Action(fname string) error {

	ofname := strings.Replace(fname, filepath.VolumeName(fname), "", -1)[1:]
	ofname = filepath.Join(h.BackupDir,
		ofname,
		time.Now().Format("060102150405")+filepath.Ext(fname))

	fnameAbs, err := filepath.Abs(fname)
	if err != nil {
		return err
	}
	ofnameAbs, err := filepath.Abs(ofname)
	if err != nil {
		return err
	}
	log.Println("backup:", fnameAbs, "->", ofnameAbs)

	// create backup dir
	if err := os.MkdirAll(filepath.Dir(ofnameAbs), 0777); err != nil {
		return err
	}

	// check outputs
	files, err := ioutil.ReadDir(filepath.Dir(ofnameAbs))
	if err != nil {
		return err
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		paths = append(paths, file.Name())
	}

	sort.Strings(paths)
	if len(files) > 0 && len(files) >= h.HistoryCount {
		rfilepath := filepath.Join(filepath.Dir(ofnameAbs), paths[0])
		log.Println("Deletefile: ", rfilepath)
		os.Remove(rfilepath)
	}

	// open input file
	fi, err := os.Open(fname)
	if err != nil {
		return err
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			log.Println("input file close Error:", err)
			return
		}
	}()
	// make a read buffer
	r := bufio.NewReader(fi)

	// open output file
	fo, err := os.Create(ofnameAbs)
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			log.Println("output file close Error:", err)
			return
		}
	}()

	// make a write buffer
	w := bufio.NewWriter(fo)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return nil
}
