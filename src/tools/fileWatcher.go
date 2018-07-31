package tools

import (
	"github.com/fsnotify"
	"path/filepath"
	"os"
	"log"
	"sync"
)

type FileWatcher struct {
	watch *fsnotify.Watcher

	mutex sync.Mutex
	watchPath map[string]bool

	fileCreateEvent chan string
	fileWriteEvent  chan string
	fileRemoveEvent chan string
	fileRenameEvent chan string
	fileChmodEvent  chan string

	//eventDone chan bool
}

func NewFileWatcher(watcher *fsnotify.Watcher,
		fileCreateEvent, fileWriteEvent, fileRemoveEvent, fileRenameEvent, fileChmodEvent chan string) *FileWatcher {

	fileWatcher := FileWatcher{}

	fileWatcher.watch = watcher
	fileWatcher.fileCreateEvent = fileCreateEvent
	fileWatcher.fileWriteEvent  = fileWriteEvent
	fileWatcher.fileRemoveEvent = fileRemoveEvent
	fileWatcher.fileRenameEvent = fileRenameEvent
	fileWatcher.fileChmodEvent  = fileChmodEvent

	fileWatcher.watchPath = make(map[string]bool)
	//fileWatcher.eventDone = make(chan bool)

	return &fileWatcher
}

func (w *FileWatcher) CleanWatchDir() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for path, _ := range w.watchPath {
		w.watch.Remove(path)
	}
	w.watchPath = map[string]bool{}
}

func (w *FileWatcher) RemoveDir(path string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.watch.Remove(path)
	delete(w.watchPath, path)
}

// Watch all directory
func (w *FileWatcher) AddwatchAllDir(dir string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Walk all directory
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Just watch directory (all child be watched)
		if info != nil && info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				log.Printf("Fail to Get Abs Path, error:%s", err.Error())
				return err
			}

			err = w.watch.Add(path)
			if err != nil {
				log.Printf("Fail to Add Watch Path, error:%s", err.Error())
				return err
			}

			log.Printf("File Watch Path: %s", path)
			w.watchPath[path] = true
		}

		return nil
	})
}

// Watch a directory
func (w *FileWatcher) AddwatchDir(dir string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	path, err := filepath.Abs(dir)
	if err != nil {
		log.Printf("Fail to Get Abs Path, error:%s", err.Error())
		return err
	}

	err = w.watch.Add(path)
	if err != nil {
		log.Printf("Fail to Add Watch Path, error:%s", err.Error())
		return err
	}

	log.Printf("Succ File Watch Path: %s", path)
	w.watchPath[path] = true

	return nil
}

func (w *FileWatcher) RunEventHandler() {
	go w.eventHandler()

	// Await
	//<-w.eventDone
}

// Handle the watch events
func (w *FileWatcher)eventHandler() {
	for {
		select {
		case ev := <-w.watch.Events:
			{
				// create event
				if ev.Op & fsnotify.Create == fsnotify.Create {
					log.Print("receive fs notify create event")

					w.fileCreateEvent <- ev.Name
					log.Printf("receive fs notify, name:%s", ev.Name)

					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						w.AddwatchDir(ev.Name)
					}
				}

				// write event
				if ev.Op & fsnotify.Write == fsnotify.Write {
					// Write事件太多，不输出日志
					w.fileWriteEvent <- ev.Name
				}

				// delete event
				if ev.Op & fsnotify.Remove == fsnotify.Remove {
					log.Print("receive fs notify remove event")

					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						w.RemoveDir(ev.Name)
					}

					w.fileRemoveEvent <- ev.Name
				}

				// rename event
				if ev.Op & fsnotify.Rename == fsnotify.Rename {
					log.Print("receive fs notify rename event")

					w.RemoveDir(ev.Name)
					w.fileRenameEvent <- ev.Name
				}

				// chmod event
				if ev.Op & fsnotify.Chmod == fsnotify.Chmod {
					log.Print("receive fs notify chmod event")

					w.fileChmodEvent <- ev.Name
				}
			}
		case err := <-w.watch.Errors:
			{
				log.Printf("fs notify occur error, error:%s", err.Error())
				//w.eventDone <- true
				return
			}
		}
	}

	//w.eventDone <- true
}

