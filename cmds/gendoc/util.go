package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"errors"
)

func detectSource() (string, error) {
	enArgs := os.Getenv("GOPATH")
	maybePaths := strings.Split(enArgs, ";")
	for index, val := range maybePaths {
		maybePaths[index] = strings.Trim(val, " ")
	}
	for _, p := range maybePaths {
		fullPath := filepath.Join(p, MidPath)
		state, err := os.Stat(fullPath)
		if err == nil {
			if state.IsDir() {
				return p, nil
			}
		}
	}
	return "", errors.New("not download source")
}

func getAllPackage(dir string) []string {
	files, _ := ioutil.ReadDir(dir)
	packagePaths := []string{}
	fileNodes := wrapfile(files, "")
	count := len(fileNodes)
	for i := 0; i < count; i++ {
		filenode := fileNodes[i]
		p := filepath.Join(filenode.parentPath, filenode.name)
		packagePaths = append(packagePaths, p)

		newPath := filepath.Join(dir, p)
		files, _ := ioutil.ReadDir(newPath)
		fileNodeTemps := wrapfile(files, p)
		count+= len(fileNodeTemps)
		fileNodes = append(fileNodes,fileNodeTemps... )
	}
	return packagePaths
}

func wrapfile(files []os.FileInfo,pPath string) []fileNode {
	fileNodes := []fileNode{}
	for _, file := range files {
		if file.IsDir() && file.Name()[0] != '.' {
			fileNodes = append(fileNodes, fileNode{
				file:       file,
				parentPath: pPath ,
				isDir:      file.IsDir(),
				name:       file.Name(),
			})
		}
	}
	return fileNodes
}

type fileNode struct {
	file       os.FileInfo
	parentPath string
	isDir      bool
	name       string
}