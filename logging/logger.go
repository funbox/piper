package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v9/fmtutil"
	"pkg.re/essentialkaos/ek.v9/timeutil"

	"github.com/gongled/piper/handler"
)

// ////////////////////////////////////////////////////////////////////////////////// //

//
type FileLogSize uint64

//
type FileLogger struct {
	w               handler.FileHandler //
	logOutput       string              //
	useTimestamp    bool                //
	maxTimeInterval int64               //
	maxBackupIndex  int                 //
	maxFileSize     FileLogSize         //
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
func Run() error {
	return Global.Run()
}

//
func SetOutput(logOutput string) {
	Global.SetOutput(logOutput)
}

//
func SetMaxBackupIndex(maxBackupIndex int) {
	Global.SetMaxBackupIndex(maxBackupIndex)
}

//
func ParseMaxTimeInterval(maxTimeInterval string) {
	Global.ParseMaxTimeInterval(maxTimeInterval)
}

//
//func SetMaxTimeInterval(maxTimeInterval int64) {
//	Global.SetMaxTimeInterval(maxTimeInterval)
//}

//
func ParseMaxFileSize(maxFileSize string) {
	Global.ParseMaxFileSize(maxFileSize)
}

//
//func SetMaxFileSize(maxFileSize FileLogSize) {
//	Global.SetMaxFileSize(maxFileSize)
//}

//
func SetTimestampFlag(flag bool) {
	Global.SetTimestampFlag(flag)
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
func (l *FileLogger) SetOutput(logOutput string) {
	l.logOutput = logOutput
}

//
func (l *FileLogger) GetOutput() string {
	return l.logOutput
}

//
func (l *FileLogger) SetMaxBackupIndex(maxBackupIndex int) {
	l.maxBackupIndex = maxBackupIndex
}

//
func (l *FileLogger) GetMaxBackupIndex() int {
	return l.maxBackupIndex
}

//
func (l *FileLogger) ParseMaxTimeInterval(maxTimeInterval string) {
	l.SetMaxTimeInterval(timeutil.ParseDuration(maxTimeInterval))
}

//
func (l *FileLogger) SetMaxTimeInterval(maxTimeInterval int64) {
	l.maxTimeInterval = maxTimeInterval
}

//
func (l *FileLogger) GetMaxTimeInterval() int64 {
	return l.maxTimeInterval
}

//
func (l *FileLogger) ParseMaxFileSize(maxFileSize string) {
	l.SetMaxFileSize(FileLogSize(fmtutil.ParseSize(maxFileSize)))
}

//
func (l *FileLogger) SetMaxFileSize(maxFileSize FileLogSize) {
	l.maxFileSize = maxFileSize
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

// ////////////////////////////////////////////////////////////////////////////////// //

//
func (l *FileLogger) Run() error {
	if l.logOutput == "" {
		return fmt.Errorf("log output must be set")
	}

	return l.w.Set(l.logOutput, 0644, l.maxTimeInterval)
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
	return (int64(time.Now().Unix()) > l.w.ExpirationTime()) && (l.w.ExpirationTime() != 0)
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
func (l *FileLogger) RenameLog() error {
	newPath := fmt.Sprintf("%s.%v", l.w.Path(), time.Now().UnixNano())

	return os.Rename(l.w.Path(), newPath)
}

//
func (l *FileLogger) RemoveStaleLogs() {
	rolledLogs := l.GetRolledOverLogs()
	sort.Sort(oldestLogFirst(rolledLogs))

	if len(rolledLogs) > l.maxBackupIndex {
		for _, staleFile := range rolledLogs[:len(rolledLogs)-l.maxBackupIndex] {
			os.Remove(staleFile)
		}
	}
}

//
func (l *FileLogger) RollOver() error {
	l.w.Close()

	if err := l.RenameLog(); err != nil {
		return err
	}

	if l.GetMaxBackupIndex() > 0 || l.GetMaxTimeInterval() > 0 {
		l.RemoveStaleLogs()
	}

	return l.w.Reopen()
}
