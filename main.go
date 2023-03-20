package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<directory>")
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	updateChan := make(chan bool, 1)

	err = watcher.Add(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	directoryTree := make(map[string][]string)

	go func() {
		err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				parentDir := filepath.Dir(path)
				directoryTree[parentDir] = append(directoryTree[parentDir], info.Name())
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		updateChan <- true
		for {
			<-updateChan
			directoryTree = make(map[string][]string)
			err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					parentDir := filepath.Dir(path)
					directoryTree[parentDir] = append(directoryTree[parentDir], info.Name())
				}
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
			updateChan <- true
		}
	}()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename ||
				event.Op&fsnotify.Chmod == fsnotify.Chmod {
				updateChan <- true
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		case <-updateChan:
			fmt.Println("Directory tree updated:")
			for parentDir, files := range directoryTree {
				fmt.Println(parentDir)
				for _, file := range files {
					fmt.Println("\t", file)
				}
			}
		}
	}
}
