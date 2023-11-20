package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

func getFileList(directory string) ([]string, error) {
	var fileList []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.Contains(path, ".git/") {
			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			fileList = append(fileList, relativePath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func deleteNonexistentFiles(directory string, fileList []string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {

			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}

			target, err := os.Readlink(path)
			if err != nil {
				log.Fatalln(err)
			}

			_, err = os.Stat(target)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.RemoveAll(path)
					if err != nil {
						log.Fatalln(err)
					}
				} else {
					log.Fatalln(err)
				}
			}

			found := false
			for _, file := range fileList {
				if strings.HasPrefix(file, relativePath) {
					found = true
					break
				}
			}

			if !found && isOriginatingFromRepo(directory, relativePath) {
				err := os.RemoveAll(path)
				if err != nil {
					return err
				}
				fmt.Printf("Deleted: %s\n", relativePath)
			}

		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func isOriginatingFromRepo(repoPath, relativePath string) bool {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	commits, err := r.Log(&git.LogOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for {
		commit, err := commits.Next()
		if err != nil {
			break
		}

		tree, err := commit.Tree()
		if err != nil {
			log.Fatal(err)
		}

		if _, err := tree.File(relativePath); err == nil {
			return true
		}

		if _, err := tree.Tree(relativePath); err == nil {
			return true
		}
	}

	return false
}
