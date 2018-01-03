
package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
	"os"
	"strings"
	"regexp"
	"fmt"
)

const MAX_OPEN_DIRS = 500


func watchDirectoryRecursive(dirToWatch string) {
	watcher, err := fsnotify.NewWatcher()
	watcherCount := 0;
	errHandler(err)
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)

				if event.Op&fsnotify.Remove != fsnotify.Remove {

					fileStat, statErr := os.Stat(event.Name)

					if statErr != nil {
						log.Println(statErr)
					}

					if event.Op&fsnotify.Create == fsnotify.Create {

						if fileStat.IsDir() {
							watcherCount = watchRecurseDirs(event.Name, watcher, watcherCount)
						} else {
							newFileNotify(event.Name)
						}
					}

					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
					}
				}
				// watchers are freed when what they are watching is removed

			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()


	watcherCount += watchRecurseDirs(dirToWatch, watcher, watcherCount)
	<-done
}

func newFileNotify(name string) {
	matchName := []string{
		"/private/tmp/foo.sh",
	}
	for _, rx := range matchName {
		matched, err := regexp.MatchString(rx, name)
		errHandler(err)
		if matched {
			fmt.Printf("New matching file: %s\n", name)
			return
		}
	}
}

func watchRecurseDirs(dirToWatch string, watcher *fsnotify.Watcher, watcherCount int) (int){

	if watchDir(dirToWatch) {
		if watcherCount+ 1 > MAX_OPEN_DIRS {
			log.Printf("Max open directories reached: %d. Not watching: %s", MAX_OPEN_DIRS, dirToWatch)
			return watcherCount
		}

		fmt.Printf("Watching directory: %s (%d)\n", dirToWatch, watcherCount)
		err := watcher.Add(dirToWatch)
		errHandler(err)
		watcherCount++;

		dirHandle, err := os.Open(dirToWatch)
		defer dirHandle.Close()
		errHandler(err)

		dirs, err := dirHandle.Readdirnames(0)
		for i := 0; i < len(dirs); i++ {
			nextDir := s.Join([]string{dirToWatch, dirs[i]},"/")
			stat, err := os.Stat(nextDir)
			errHandler(err)
			if stat.IsDir() {
				watcherCount = watchRecurseDirs(nextDir, watcher, watcherCount)
			}

		}
	}

	return watcherCount;
}

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func watchDir (d string) (bool) {
	excludeStrings := []string{
		".rbenv",
		".rvm",
		".kitchen",
		".vagrant.d",
		".Trash",
		".DS_Store",
		"Library",
		"Downloads",
		"miniconda",
		"go",
		"uuid",
		"sdk",
		"cache",
	}

	for _, w := range excludeStrings {
		if strings.Contains(d, w) {
			return false
		}
	}
	return true
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please supply a base directory name")
	}
	startDir := os.Args[1]
	stat, err := os.Stat(startDir)
	errHandler(err);
	if stat.IsDir() {
		watchDirectoryRecursive(os.Args[1])
	}
}