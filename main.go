package main

import (
	"fmt"
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
	spath         string
	dpath         string
	filePattern   string
	keepIfExisted bool
)

func main() {
	flag.StringVar(&spath, "spath", "./", "search source path")
	flag.StringVar(&dpath, "dpath", "./", "destination path")
	flag.StringVar(&filePattern, "file", "*", "file pattern")
	//flag.BoolVar(&keepIfExisted, "keep", true, "keep with a new name if filename exists")
	flag.Parse()
	filePattern = strings.ToLower(filePattern)
	filepath.Walk(spath, visit)
}

func visit(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}
	fileName := f.Name()
	fileFullName := path
	if ok, err := filepath.Match(filePattern, strings.ToLower(fileName)); !ok || err != nil {
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
			if !keepIfExisted {
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
