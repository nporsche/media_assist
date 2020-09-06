package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"flag"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

var (
	spath          string
	dpath          string
	filePatternRaw string
	filePattern    []string
)

func main() {
	flag.StringVar(&spath, "spath", "./", "search source path")
	flag.StringVar(&dpath, "dpath", "./", "destination path")
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
		log.Println("WARNING: not match pattern", f.Name())
		return nil
	}

	tm, err := getExifDatetime(fileFullName)
	if err != nil {
		log.Printf("WARNING: cannot get exif info from %s use mod time instead\n", fileFullName)
		tm = f.ModTime()
	}
	destinateDir := filepath.Join(dpath, fmt.Sprintf("%4d", tm.Year()), fmt.Sprintf("%02d", int(tm.Month())))

	_, err = exec.Command("mkdir", "-p", destinateDir).Output()
	if err != nil {
		log.Printf("FAILED: mkdir -p %s\n", destinateDir, err)
		return nil
	}

	var targetFileName string
	var targetFileFullName string

	for i := 0; ; i++ {
		if _, err := os.Stat(fileFullName); err == nil {
			//source file is not moved
			if i == 0 {
				targetFileName = tm.Format("2006-01-02 150405") + filepath.Ext(f.Name())
			} else {
				targetFileName = fmt.Sprintf("%s(%d)%s", tm.Format("2006-01-02 150405"), i, filepath.Ext(f.Name()))
			}
			targetFileFullName = filepath.Join(destinateDir, targetFileName)
			_, err = exec.Command("mv", "-n", fileFullName, targetFileFullName).Output()
			if err != nil {
				log.Printf("FAILED: mv -n %s %s\n", fileFullName, targetFileFullName)
				break
			}
			d1 := digest(targetFileFullName)
			d2 := digest(fileFullName)
			if d1 == d2 {
				log.Printf("INFO: IGNORE mv %s %s because same digest %s\n", fileFullName, targetFileFullName, d1)
				break
			}
		} else {
			break
		}
	}
	log.Printf("SUCCESS: mv -n %s %s\n", fileFullName, targetFileFullName)

	return nil
}

func getExifDatetime(fname string) (tm time.Time, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return tm, err
	}

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil {
		return tm, err
	}

	// Two convenience functions exist for date/time taken and GPS coords:
	return x.DateTime()
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
