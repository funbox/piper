package log

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gongled/piper/handler"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// FileLogger main struct for logger
type FileLogger struct {
	w               handler.FileHandler //
	r               *os.File            //
	logOutput       string              //
	useTimestamp    bool                //
	maxTimeInterval int64               //
	maxBackupIndex  int                 //
	maxFileSize     uint64              //
}

// Logging time format
const FILE_LOGGER_FORMAT = "02/Jan/2006:15:04:05"

// Sort interface
type oldestLogFirst []string

// ////////////////////////////////////////////////////////////////////////////////// //

// Global instance of logger
var Global = &FileLogger{}

// ////////////////////////////////////////////////////////////////////////////////// //

// Len length of the slice
func (s oldestLogFirst) Len() int {
	return len(s)
}

// Swap swap elements between each other
func (s oldestLogFirst) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less compares elements and returns the lower one
func (s oldestLogFirst) Less(i, j int) bool {
	return getRotateFileSuffix(s[i]) < getRotateFileSuffix(s[j])
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getRotateFileSuffix gets numeric suffix for rotating files
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

// ////////////////////////////////////////////////////////////////////////////////// //

// Run runs logger
func Run() error {
	return Global.Run()
}

// SetInput set the input
func SetInput(rd *os.File) {
	Global.SetInput(rd)
}

// SetOutput set the output
func SetOutput(logOutput string) {
	Global.SetOutput(logOutput)
}

// SetMaxBackupIndex sets up maximum amount of the rotating file
func SetMaxBackupIndex(maxBackupIndex int) {
	Global.SetMaxBackupIndex(maxBackupIndex)
}

// SetMaxTimeInterval sets up duration for keeping files up
func SetMaxTimeInterval(maxTimeInterval int64) {
	Global.SetMaxTimeInterval(maxTimeInterval)
}

// SetMaxFileSize sets up maximum file size for keeping files up
func SetMaxFileSize(maxFileSize uint64) {
	Global.SetMaxFileSize(maxFileSize)
}

// SetTimestampFlag says append timestamp before log line or not
func SetTimestampFlag(flag bool) {
	Global.SetTimestampFlag(flag)
}

// RollOver rotates file
func RollOver() error {
	return Global.RollOver()
}

// Close closes log
func Close() error {
	return Global.w.Close()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// SetInput set the input
func (l *FileLogger) SetInput(rd *os.File) {
	l.r = rd
}

// SetOutput set the output
func (l *FileLogger) SetOutput(logOutput string) {
	l.logOutput = logOutput
}

// SetMaxBackupIndex sets up maximum amount of the rotating file
func (l *FileLogger) SetMaxBackupIndex(maxBackupIndex int) {
	l.maxBackupIndex = maxBackupIndex
}

// GetMaxBackupIndex gets maximum amount of the rotating files
func (l *FileLogger) GetMaxBackupIndex() int {
	return l.maxBackupIndex
}

// SetMaxTimeInterval sets up duration for keeping files up
func (l *FileLogger) SetMaxTimeInterval(maxTimeInterval int64) {
	l.maxTimeInterval = maxTimeInterval
}

// GetMaxTimeInterval gets duration for keeping files up
func (l *FileLogger) GetMaxTimeInterval() int64 {
	return l.maxTimeInterval
}

// SetMaxFileSize sets up maximum file size for keeping files up
func (l *FileLogger) SetMaxFileSize(maxFileSize uint64) {
	l.maxFileSize = maxFileSize
}

// GetMaxFileSize gets maximum file size for keeping files up
func (l *FileLogger) GetMaxFileSize() uint64 {
	return l.maxFileSize
}

// SetTimestampFlag says append timestamp before log line or not
func (l *FileLogger) SetTimestampFlag(flag bool) {
	l.useTimestamp = flag
}

// GetTimestampFlag returns should we append timestamp before log line or not
func (l *FileLogger) GetTimestampFlag() bool {
	return l.useTimestamp
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Run runs the logger
func (l *FileLogger) Run() error {
	if l.logOutput == "" {
		return fmt.Errorf("log output must be set")
	}

	if err := l.w.Set(l.logOutput, 0644, l.maxTimeInterval); err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if err := l.WriteLog(scanner.Text()); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// timeNow returns current time in the custom format
func (l *FileLogger) timeNow() string {
	t := time.Now()

	return fmt.Sprintf("%s.%09d", t.Format(FILE_LOGGER_FORMAT), t.Nanosecond())
}

// prependTimestamp prepends timestamp to entry
func (l *FileLogger) prependTimestamp(entry string) string {
	return fmt.Sprintf("[%s] %s", l.timeNow(), entry)
}

// Append appends line to the writer
func (l *FileLogger) Append(p []byte) {
	l.w.Write(p)
}

// AppendLine appends line with a CRLF to the writer
func (l *FileLogger) AppendLine(p []byte) {
	l.Append(p)
	l.Append([]byte{'\n'})
}

// FormatEntry formats entry before writing to the log
func (l *FileLogger) FormatEntry(entry string) string {
	if l.GetTimestampFlag() {
		entry = l.prependTimestamp(entry)
	}

	return entry
}

// WriteLog writes formatted entry to the log
func (l *FileLogger) WriteLog(entry string) error {
	entry = l.FormatEntry(entry)

	if l.IsMaxFileSizeReached(entry) || l.IsMaxFileAgeReached() {
		if err := l.RollOver(); err != nil {
			return err
		}
	}

	l.AppendLine([]byte(entry))

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// IsMaxFileSizeReached checks if maximum file size is reached
func (l *FileLogger) IsMaxFileSizeReached(entry string) bool {
	// TODO: fix types casting
	return (l.GetMaxFileSize() != 0) && (l.w.Size()+uint64(len(entry)) >= uint64(l.GetMaxFileSize()))
}

// isMaxFileAgeReached checks if maximum file age is reached
func (l *FileLogger) IsMaxFileAgeReached() bool {
	return (int64(time.Now().Unix()) > l.w.ExpirationTime()) && (l.w.ExpirationTime() != 0)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetRolledOverLogs returns list of rotated logs
func (l *FileLogger) GetRolledOverLogs() []string {
	globPattern := fmt.Sprintf("%s.*", l.w.Path())
	rolledLogs, err := filepath.Glob(globPattern)

	if err != nil {
		return nil
	}

	return rolledLogs
}

// RenameLog renames log file
func (l *FileLogger) RenameLog() error {
	newPath := fmt.Sprintf("%s.%v", l.w.Path(), time.Now().UnixNano())

	return os.Rename(l.w.Path(), newPath)
}

// RemoveStaleLogs removes staled logs
func (l *FileLogger) RemoveStaleLogs() {
	rolledLogs := l.GetRolledOverLogs()
	sort.Sort(oldestLogFirst(rolledLogs))

	if len(rolledLogs) > l.maxBackupIndex {
		for _, staleFile := range rolledLogs[:len(rolledLogs)-l.maxBackupIndex] {
			os.Remove(staleFile)
		}
	}
}

// RollOver rotates logs
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
