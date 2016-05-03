/*
---
author: qinguoan@wandoujia.com
*/
package tail

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats"
	"gopkg.in/tomb.v1"
	"io"
	"io/ioutil"
	"log"
	"modi/taild/tail/watch"
	"os"
	"path"
	"strings"
	"time"
)

var (
	// DefaultLogger is used when Config.Logger == nil
	DefaultLogger = log.New(os.Stdout, "", log.LstdFlags)
	// DiscardingLogger can be used to disable logging output
	DiscardingLogger = log.New(ioutil.Discard, "", 0)
	// stop tail
	ErrStop = fmt.Errorf("tail should now stop")
)

type NewLine struct {
	HostName string
	FileName string
	LineText string
}

type Tail struct {
	tomb.Tomb        // provides: Done, Kill, Dying
	FileName  string // log file name.
	host      string
	topic     string
	file      *os.File      // log file desc.
	reader    *bufio.Reader // reader for lof gile.
	watcher   *watch.PollingFileWatcher
	changes   *watch.FileChanges
	NC        *nats.Conn
}

func (tail *Tail) reopen() (err error) {
	if tail.file != nil {
		tail.file.Close()
	}
	if tail.reader != nil {
		tail.reader = nil
	}
	for {
		tail.file, err = os.Open(tail.FileName)
		if err != nil {
			if os.IsNotExist(err) {
				if err := tail.watcher.BlockUntilExists(&tail.Tomb); err != nil {
					if err == tomb.ErrDying {
						return err
					}
					return fmt.Errorf("Failed to detect creation of %s: %s",
						tail.FileName, err)
				}
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			return fmt.Errorf("Unable to open file %s: %s", tail.FileName, err)
		}
		tail.reader = nil
		tail.reader = bufio.NewReader(tail.file)
		break
	}
	return
}

func (tail *Tail) clean() {
	tail.Done()
	if tail.file != nil {
		tail.file.Close()
	}
}

func (tail *Tail) seekEnd() error {
	_, err := tail.file.Seek(0, 2)
	if err != nil {
		return fmt.Errorf("Seek error on %s: %s", tail.FileName, err)
	}
	tail.reader.Reset(tail.file)
	return nil
}

func (tail *Tail) waitForChange() error {
	if tail.changes == nil {
		st, err := tail.file.Stat()
		if err != nil {
			return err
		}
		tail.changes = tail.watcher.ChangeEvents(&tail.Tomb, st)
	}
	select {
	case <-tail.changes.Modified:
		return nil
	case <-tail.changes.Deleted:
		tail.changes = nil
		if err := tail.reopen(); err != nil {
			return err
		}
		return nil
	case <-tail.changes.Truncated:
		if err := tail.reopen(); err != nil {
			return err
		}
		return nil
	case <-tail.Dying():
		return ErrStop
	}
	return nil
}

func (tail *Tail) readLine() (string, error) {
	line, err := tail.reader.ReadString('\n')
	if err != nil {
		return line, err
	}

	line = strings.TrimRight(line, "\n")
	return line, err
}

func (tail *Tail) tailFileSync() {
	defer tail.clean()
	err := tail.reopen()
	if err != nil {
		tail.Kill(err)
	}
	err = tail.seekEnd()
	if err != nil {
		tail.Kill(err)
	}

	var line string
	var data []byte
	var newline *NewLine

	for {

		line, err = tail.readLine()

		if err == nil || (err == io.EOF && line != "") {
			newline = &NewLine{
				HostName: tail.host,
				FileName: tail.FileName,
				LineText: line,
			}
			data, _ = json.Marshal(newline)
			tail.NC.Publish(tail.topic, data)
		} else if err == io.EOF {
			err := tail.waitForChange()
			if err != nil {
				if err != ErrStop {
					tail.Kill(err)
				}
				return
			}
		} else {
			tail.Killf("Error reading %s: %s", tail.FileName, err)
			return
		}

		select {
		case <-tail.Dying():
			return
		default:
		}
	}
}

func NewTail(filename string, nc *nats.Conn) (*Tail, error) {
	hostname, _ := os.Hostname()
	fn := path.Base(filename)
	fn = strings.Replace(fn, ".", "_", -1)
	prefix := strings.Split(fn, "_")
	t := &Tail{
		FileName: filename,
		host:     hostname,
		topic:    prefix[0],
		NC:       nc,
	}
	t.watcher = watch.NewPollingFileWatcher(filename)
	go t.tailFileSync()
	return t, nil
}
