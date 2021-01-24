package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func GetFilesInFolder(root string, ext string) []string {
	var files []string
	logrus.Debugf("looking for files with extension %s in %s\n", root, ext)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		logrus.Debugf("path: %s\n", path)
		if filepath.Ext(path) == ext {
			logrus.Debugf("add %s\n", path)
			files = append(files, strings.Replace(path, root, "", 1))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		logrus.Debug(file)
	}
	return files
}

func ReadFileLines(path string) (res []string) {
	file, err := os.Open(path)
	if err != nil {
		logrus.Error(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logrus.Error(err)
	}
	return res
}
