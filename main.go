package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func fsnotifyHandler(watcher *fsnotify.Watcher, handler Handler) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("fatal: watcher.Events got not ok flg.")
				return
			}
			// Check dir and add serch path
			fInfo, err := os.Stat(event.Name)
			if err != nil {
				if fInfo == nil {
					log.Println("DirRmv - " + event.Name)
					watcher.Remove(event.Name)
					continue
				} else {
					log.Println("Err: " + err.Error())
				}
			} else if fInfo.IsDir() {
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("DirAdd - " + event.Name)
					watcher.Add(event.Name)
					continue
				}
			}
			if !handler.CheckTarget(event.Name) {
				continue
			}
			// Optimezed for windows.
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Action:", event.Name)
				handler.Action(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("fatal: watcher.Errors got not ok flg.")
				return
			}
			log.Println("error:", err)
		}
	}
}

func main() {

	logfile, err := os.OpenFile("backuper.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open backuper.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	flag.Parse()
	log.Println(flag.Arg(0))
	conffile := flag.Arg(0)

	var conf YamlConfLoader
	conf.Load(conffile)

	var hdl EmyaHandler
	err = hdl.Init(&conf)
	if err != nil {
		log.Fatal("Handler init failed")
	}

	_main(&conf, &hdl)
}

func _main(conf ConfigLoader, hdl Handler) {
	targets := conf.GetTgts()
	var tgtlist []string
	for _, fname := range targets {
		if !Exists(fname) {
			log.Println("Directory not exists :", fname)
			continue
		}
		tgtlist = append(tgtlist, fname)
	}
	log.Println("targets : ", tgtlist)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go fsnotifyHandler(watcher, hdl)

	for _, fname := range tgtlist {
		err := filepath.Walk(fname, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() {
				err = watcher.Add(p)
				if err != nil {
					log.Println("watcher.Add failed : ", p, " - ", err)
				}
			}
			return nil
		})
		if err != nil {
			log.Println("watcher.Walk failed : ", fname, " - ", err)
		}
		// err = watcher.Add(fname)
		// if err != nil {
		// 	log.Println("watcher.Add failed : ", fname, " - ", err)
		// }
	}
	<-done
}
