package main

import (
	"sync"
	"log"
	"config"
	"tools"
	"github.com/fsnotify"
	"time"
)

type filenotifyLogUploader struct {
	task      string
	cfg       *config.CliJsonConfig
	uploader  *uploader

	fileWatcher *tools.FileWatcher

	wg  sync.WaitGroup
	// create file
	fileCreateEvent	chan string
	// write file
	fileWriteEvent	chan string
	// remove file
	fileRemoveEvent	chan string
	// rename file
	fileRenameEvent	chan string
	// chmod file
	fileChmodEvent	chan string

	isOver bool
}

func NewFileNotifyLogUploader(task string, cfg *config.CliJsonConfig, uploader *uploader) *filenotifyLogUploader {
	logUploader := &filenotifyLogUploader{}
	logUploader.task     = task
	logUploader.cfg      = cfg
	logUploader.uploader = uploader

	logUploader.fileCreateEvent = make(chan string)
	logUploader.fileWriteEvent  = make(chan string)
	logUploader.fileRemoveEvent = make(chan string)
	logUploader.fileRenameEvent = make(chan string)
	logUploader.fileChmodEvent  = make(chan string)

	watcher, _ := fsnotify.NewWatcher()
	logUploader.fileWatcher = tools.NewFileWatcher(watcher,
		logUploader.fileCreateEvent, logUploader.fileWriteEvent,
		logUploader.fileRemoveEvent, logUploader.fileRenameEvent, logUploader.fileChmodEvent)

	logUploader.isOver = false

	return logUploader
}

func (u *filenotifyLogUploader) Start() {
	log.Printf("file notify log uploader start, taskname:%s", u.task)

	if ok, logDir := u.cfg.GetUploadLogLogDir(u.task); !ok {
		log.Printf("file notify log uploader game over, taskname:%s", u.task)
		return
	} else {
		if err := u.fileWatcher.AddwatchDir(logDir); err != nil {
			log.Printf("file notify log uploader game over, taskname:%s", u.task)
			return
		}
	}

	u.fileWatcher.RunEventHandler()  // TODO

	u.wg.Add(5)

	// create event handler
	go func() {
		for {
			select {
			case createFileName := <-u.fileCreateEvent: // TODO 不监控崩溃日志
				log.Print("createFile: " + createFileName)

				ok, logServerUrl := u.cfg.GetUploadLogLogServerUrl(u.task)
				if !ok {
					log.Printf("file notify log uploader, create file game over, taskname:%s", u.task)
					u.isOver = true
					return
				}

				u.uploader.LogNotify(createFileName, logServerUrl, 0)
			case <- time.After(30 * time.Second):
				if ok, _ := u.cfg.GetUploadLogLogDir(u.task); !ok {
					log.Printf("file notify log uploader, create file game over, taskname:%s", u.task)
					u.isOver = true
					return
				}
			}
		}

		defer u.wg.Done()
	}()

	// write event handler
	go func() {
		for {
			select {
			case <-u.fileWriteEvent:
				{
				}
			case <- time.After(30 * time.Second):
				{
					if u.isOver {
						log.Printf("file notify log uploader, write file game over, taskname:%s", u.task)
						return
					}
				}
			}
		}

		defer u.wg.Done()
	}()

	// remove event handler
	go func() {
		for {
			select {
			case removeFileName := <-u.fileRemoveEvent:
				{
					log.Print("remove file: " + removeFileName)
				}
			case <- time.After(30 * time.Second):
				{
					if u.isOver {
						log.Printf("file notify log uploader, remove file game over, taskname:%s", u.task)
						return
					}
				}
			}
		}

		defer u.wg.Done()
	}()

	// rename event handler
	go func() {
		for {
			select {
			case renameFileName := <-u.fileRenameEvent:
				{
					log.Print("rename file: " + renameFileName)
				}
			case <- time.After(30 * time.Second):
				{
					if u.isOver {
						log.Printf("file notify log uploader, rename file game over, taskname:%s", u.task)
						return
					}
				}
			}
		}

		defer u.wg.Done()
	}()

	// chmod event handler
	go func() {
		for {
			select {
			case chmodFileName := <-u.fileChmodEvent:
				{
					log.Print("chmod file: " + chmodFileName)
				}
			case <- time.After(30 * time.Second):
				{
					if u.isOver {
						log.Printf("file notify log uploader, chmod file game over, taskname:%s", u.task)
						return
					}
				}
			}
		}

		defer u.wg.Done()
	}()

	u.wg.Wait()
}
