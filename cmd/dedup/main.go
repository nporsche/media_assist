package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"flag"
)

var (
	spath          string
	dpath          string
	filePatternRaw string
	filePattern    []string

	fileDigestMapping map[string]string
)

func main() {
	flag.StringVar(&spath, "spath", "./", "search source path")
	flag.StringVar(&dpath, "delete-path", "./", "dup files deleted path")
	flag.StringVar(&filePatternRaw, "file", "*", "file pattern")
	flag.Parse()

	fileDigestMapping = make(map[string]string)

	filePattern = strings.Split(strings.ToLower(filePatternRaw), ";")
	filepath.Walk(spath, visit)
}

func visit(path string, f os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if f.IsDir() {
		return nil
	}
	fileName := f.Name()
	fileFullName := path
	match := false
	for _, pattern := range filePattern {
		if ok, err := filepath.Match(pattern, strings.ToLower(fileName)); err == nil && ok {
			match = true
			break
		}
	}
	if !match {
		//log.Println("WARNING: not match pattern", f.Name())
		return nil
	}

	md5 := digest(fileFullName)

	if _, ok := fileDigestMapping[md5]; !ok {
		fileDigestMapping[md5] = fileFullName
		return nil
	}

	fmt.Printf("%s is deplicated with %s", fileFullName, fileDigestMapping[md5])
	if fileSize(fileFullName) != fileSize(fileDigestMapping[md5]) {
		fmt.Println("Warning: Size is different but digest is same")
	}

	return nil
}

func digest(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fileSize(filename string) int64{
	info, err := os.Stat(filename)
	if err != nil{
		panic(err)
	}
	return info.Size()
}
