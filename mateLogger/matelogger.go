package matelogger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var logChannel chan string
var close bool

type LogFile struct {
	mu        sync.Mutex
	path      string
	name      string
	file      *os.File
	logDate   string
	logNum    int
	logStdout bool

	limitLogSize int
	innerLogSize int
}

func Start() {
	go Run()
}

func Run() {
	for {
		log.Print(<-logChannel)

		if close == true && len(logChannel) == 0 {
			break
		}
	}
}

func Close() {
	close = true
	for {
		if len(logChannel) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(10 * time.Millisecond)
}

func SetupLog(path string, name string, logStdout bool) *LogFile {
	log.SetFlags(0)
	if path == "" {
		path = "."
	}

	logChannel = make(chan string, 1000)

	//로그 디렉터리를 생성한다.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Logger Error : " + err.Error())
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("Logger Error : " + err.Error())
			os.Exit(2)
		}
	}

	lf, err := NewLogFile(path, name, nil, 5, logStdout)
	if err != nil {
		log.Fatalf("Unable to create log file: %s", err.Error())
	}

	log.SetOutput(lf)

	return lf
}

// NewLogFile creates a new LogFile. The file is optional - it will be created if needed.
func NewLogFile(path string, name string, file *os.File, limitLogSize int, logStdout bool) (*LogFile, error) {
	rw := &LogFile{
		file:         file,
		path:         path,
		name:         name,
		limitLogSize: limitLogSize,
		innerLogSize: 0,
		logNum:       0,
		logDate:      time.Now().Format("2006-01-02"),
		logStdout:    logStdout,
	}

	if logStdout {
		rw.file = os.Stdout
	} else if file == nil {
		if err := rw.Rotate(); err != nil {
			return nil, err
		}
	}
	return rw, nil
}

func (l *LogFile) Write(b []byte) (n int, err error) {
	if l.logStdout {
		n, err = l.file.Write(b)
		return
	}

	if l.logDate != time.Now().Format("2006-01-02") {
		l.logNum = 0
		l.Rotate()
	}

	l.mu.Lock()
	n, err = l.file.Write(b)
	l.innerLogSize += n
	l.mu.Unlock()

	if l.innerLogSize > (1024 * 1024 * l.limitLogSize) {
		l.Rotate()
	}
	return
}

func (l *LogFile) Rotate() error {
	l.mu.Lock()
	l.mu.Unlock()
	// rename dest file if it already exists.
	if _, err := os.Stat(filepath.Join(l.path, l.name+"_"+l.logDate+".log")); err == nil {
		//이전에 생성된 같은 이름의 로그를 건너뛰기 위해 파일 검사
		var name string
		for {
			name = l.name + "_" + l.logDate + "__" + strconv.Itoa(l.logNum) + ".log"
			if _, err := os.Stat(filepath.Join(l.path, name)); err != nil {
				break
			}
			l.logNum++
		}

		l.file.Close()

		if l.innerLogSize > (1024 * 1024 * l.limitLogSize) {
			if err = os.Rename(filepath.Join(l.path, l.name+"_"+l.logDate+".log"), filepath.Join(l.path, name)); err != nil {
				return err
			}
		}
	}
	l.logDate = time.Now().Format("2006-01-02")

	file, err := os.OpenFile(filepath.Join(l.path, l.name+"_"+l.logDate+".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	l.file = file

	if info, err := l.file.Stat(); err == nil {
		l.innerLogSize = int(info.Size())
	} else {
		l.innerLogSize = 0
	}

	return nil
}

func Custom(format string) {
	logChannel <- format
}
