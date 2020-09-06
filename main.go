package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"flag"
)

var (
	spath          string
	dpath          string
	unpath         string
	filePatternRaw string
	filePattern    []string
)

func main() {
	flag.StringVar(&spath, "spath", "./", "search source path")
	flag.StringVar(&dpath, "dpath", "./", "destination path")
	flag.StringVar(&unpath, "unpath", "./", "path that has unhandled files")
	flag.StringVar(&filePatternRaw, "file", "*", "file pattern")
	flag.Parse()
	filePattern = strings.Split(strings.ToLower(filePatternRaw), ";")
	filepath.Walk(spath, visit)
}

func visit(path string, f os.FileInfo, err error) error {
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

	tm, err := getExifDatetime(fileFullName)
	if err != nil {
		log.Printf("FAILED: cannot get exif info from %s and mv to %s", fileFullName, unpath)
		exec.Command("mv", fileFullName, unpath).Output()
		return nil
	}
	tm = tm.Local()
	destinateDir := filepath.Join(dpath, fmt.Sprintf("%4d", tm.Year()), fmt.Sprintf("%02d", int(tm.Month())))

	_, err = exec.Command("mkdir", "-p", destinateDir).Output()
	if err != nil {
		log.Printf("FAILED: mkdir -p %s\n", destinateDir, err)
		return nil
	}

	var targetFileName string
	var targetFileFullName string

	for i := 0; ; i++ {
		if fileExists(fileFullName) {
			//source file is not moved
			if i == 0 {
				targetFileName = tm.Format("2006-01-02 150405") + filepath.Ext(f.Name())
			} else {
				targetFileName = fmt.Sprintf("%s(%d)%s", tm.Format("2006-01-02 150405"), i, filepath.Ext(f.Name()))
			}
			targetFileFullName = filepath.Join(destinateDir, targetFileName)

			if targetFileFullName == fileFullName {
				return nil
			}

			if fileExists(targetFileFullName) {
				d1 := digest(targetFileFullName)
				d2 := digest(fileFullName)
				if d1 == d2 {
					os.Remove(fileFullName)
					log.Printf("INFO: rm %s because %s same digest %s\n", fileFullName, targetFileFullName, d1)
					return nil
				}
			} else {
				_, err = exec.Command("mv", "-n", fileFullName, targetFileFullName).Output()
				if err != nil {
					panic(err)
				}
			}
		} else {
			break
		}
	}
	log.Printf("SUCCESS: mv -n %s %s\n", fileFullName, targetFileFullName)

	return nil
}

func getExifDatetime(fname string) (exifCreateTM time.Time, err error) {
	//cmd := fmt.Sprintf(`"%s" |grep -i "Track Create Date"|awk -F ": " '{print $2}'`, fname)
	//Creation Date                   : 2018:09:16 11:44:34+08:00
	out, err := exec.Command("/usr/local/bin/exiftool", fname).Output()
	if err != nil {
		return exifCreateTM, err
	}
	var createTm *time.Time
	var fileMod *time.Time
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Create Date") {
			//Mon Jan 2 15:04:05 -0700 MST 2006
			tmString := strings.Split(line, ": ")[1]
			tm, err := time.ParseInLocation("2006:01:02 15:04:05", tmString, time.UTC)
			if err == nil {
				createTm = &tm
			}
		} else if strings.HasPrefix(line, "File Modification Date/Time") {
			tmString := strings.Split(line, ": ")[1]
			tm, err := time.Parse("2006:01:02 15:04:05-07:00", tmString)
			if err == nil {
				fileMod = &tm
			}
		}
		if createTm != nil && fileMod != nil {
			break
		}
	}
	if createTm == nil && fileMod == nil {
		return exifCreateTM, errors.New("Failed get exif info because not prefix Create Date")
	}
	if createTm != nil {
		exifCreateTM = *createTm
	} else {
		exifCreateTM = *fileMod
	}
	log.Printf("INFO: %s exif date time %v", fname, exifCreateTM)
	return exifCreateTM, nil
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
