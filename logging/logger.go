package log

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"path/filepath"

	"pkg.re/essentialkaos/ek.v9/fmtutil"

	"github.com/gongled/piper/handler"
	"pkg.re/essentialkaos/ek/timeutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

//
type FileLogSize uint64

//
type FileLogger struct {
	w              handler.FileHandler  //
	useTimestamp   bool                 //
	maxFileAge     int64                //
	maxBackupIndex int                  //
	maxFileSize    FileLogSize          //

	nextRollOverTime int64              //
}

const FILE_LOGGER_FORMAT = "02/Jan/2006:15:04:05"

//
type oldestLogFirst []string

// ////////////////////////////////////////////////////////////////////////////////// //

var Global = &FileLogger{}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func getRotateFileSuffix(s string) int64 {
	pieces := strings.Split(s, ".")

	if len(pieces) < 2 {
		return 0
	}

	suffix, err := strconv.ParseInt(pieces[len(pieces)-1], 10, 64)

	if err != nil {
		return 0
	}

	return suffix
}

//
func (s oldestLogFirst) Len() int {
	return len(s)
}

//
func (s oldestLogFirst) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//
func (s oldestLogFirst) Less(i, j int) bool {
	return getRotateFileSuffix(s[i]) < getRotateFileSuffix(s[j])
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l FileLogSize) Pretty() string {
	return fmtutil.PrettySize(l)
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func SetUp(logFile string) error {
	return Global.SetUp(logFile)
}

//
func SetMaxBackupIndex(maxBackupIndex int) {
	Global.SetMaxBackupIndex(maxBackupIndex)
}

func SetMaxFileSize(maxFileSize string) {
	Global.SetMaxFileSize(maxFileSize)
}

//
func SetTimestampFlag(flag bool) {
	Global.SetTimestampFlag(flag)
}

//
func SetMaxFileAge(maxFileAge string) {
	Global.SetMaxFileAge(maxFileAge)
}

//
func FormatEntry(entry string) string {
	return Global.FormatEntry(entry)
}

//
func WriteLog(entry string) error {
	return Global.WriteLog(entry)
}

//
func RollOver() error {
	return Global.RollOver()
}

//
func Close() error {
	return Global.w.Close()
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l *FileLogger) SetMaxBackupIndex(maxBackupIndex int) {
	l.maxBackupIndex = maxBackupIndex
}

//
func (l *FileLogger) GetMaxBackupIndex() int {
	return l.maxBackupIndex
}

//
func (l *FileLogger) SetMaxFileSize(maxFileSize string) {
	l.maxFileSize = FileLogSize(fmtutil.ParseSize(maxFileSize))
}

//
func (l *FileLogger) GetMaxFileSize() FileLogSize {
	return l.maxFileSize
}

//
func (l *FileLogger) SetTimestampFlag(flag bool) {
	l.useTimestamp = flag
}

//
func (l *FileLogger) GetTimestampFlag() bool {
	return l.useTimestamp
}

//
func (l *FileLogger) SetMaxFileAge(maxFileAge string) {
	l.maxFileAge = timeutil.ParseDuration(maxFileAge)
}

//
func (l *FileLogger) GetMaxFileAge() int64 {
	return l.maxFileAge
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l *FileLogger) SetUp(logFile string) error {
	return l.w.Set(logFile, 0644)
}

//
func (l *FileLogger) timeNow() string {
	t := time.Now()

	return fmt.Sprintf("%s.%09d", t.Format(FILE_LOGGER_FORMAT), t.Nanosecond())
}

//
func (l *FileLogger) prependTimestamp(entry string) string {
	return fmt.Sprintf("[%s] %s", l.timeNow(), entry)
}

//
func (l *FileLogger) Append(p []byte) {
	l.w.Write(p)
}

//
func (l *FileLogger) AppendLine(p []byte) {
	l.Append(p)
	l.Append([]byte{'\n'})
}

//
func (l *FileLogger) FormatEntry(entry string) string {
	if l.GetTimestampFlag() {
		entry = l.prependTimestamp(entry)
	}

	return entry
}

//
func (l *FileLogger) WriteLog(entry string) error {
	if l.IsMaxFileSizeReached(entry) || l.IsMaxFileAgeReached() {
		if err := l.RollOver(); err != nil {
			return err
		}
	}

	l.AppendLine([]byte(entry))

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l *FileLogger) IsMaxFileSizeReached(entry string) bool {
	// TODO: fix types casting
	return (l.GetMaxFileSize() != 0) && (l.w.Size()+uint64(len(entry)) >= uint64(l.GetMaxFileSize()))
}

//
func (l *FileLogger) IsMaxFileAgeReached() bool {
	return int64(time.Now().Second()) > l.nextRollOverTime

	// return false
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l *FileLogger) GetRolledOverLogs() []string {
	globPattern := fmt.Sprintf("%s.*", l.w.Path())

	rolledLogs, err := filepath.Glob(globPattern)

	if err != nil {
		return nil
	}

	return rolledLogs
}

//
func (l *FileLogger) RollOver() error {
	l.w.Close()

	newPath := fmt.Sprintf("%s.%v", l.w.Path(), time.Now().UnixNano())

	if err := os.Rename(l.w.Path(), newPath); err != nil {
		return err
	}

	if l.maxBackupIndex > 0 {
		rolledLogs := l.GetRolledOverLogs()
		sort.Sort(oldestLogFirst(rolledLogs))

		if len(rolledLogs) > l.maxBackupIndex {
			for _, staleFile := range rolledLogs[:len(rolledLogs)-l.maxBackupIndex] {
				os.Remove(staleFile)
			}
		}
	}

	return l.w.Reopen()
}
