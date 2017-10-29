package cli

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/gongled/piper/logging"

	"pkg.re/essentialkaos/ek.v9/fmtc"
	"pkg.re/essentialkaos/ek.v9/fmtutil"
	"pkg.re/essentialkaos/ek.v9/options"
	"pkg.re/essentialkaos/ek.v9/signal"
	"pkg.re/essentialkaos/ek.v9/usage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Application info
const (
	APP  = "piper"
	VER  = "1.0.0"
	DESC = "Utility for log rotation for 12-factor apps"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_SIZELIMIT = "S:size"
	OPT_TIMESTAMP = "t:timestamp"
	OPT_NO_COLOR  = "nc:no-color"
	OPT_HELP      = "h:help"
	OPT_VERSION   = "v:version"
)

const TS_PIPER_FORMAT = "02/Jan/2006:15:04:05.999999 -07:00"

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_SIZELIMIT: {Type: options.STRING},
	OPT_TIMESTAMP: {Type: options.BOOL},
	OPT_NO_COLOR:  {Type: options.BOOL},
	OPT_HELP:      {Type: options.BOOL, Alias: "u:usage"},
	OPT_VERSION:   {Type: options.BOOL, Alias: "ver"},
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Init provides entry point to the program
func Init() {
	opts, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		shutdown()
	}

	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if options.GetB(OPT_VERSION) {
		showAbout()
		os.Exit(0)
	}

	if options.GetB(OPT_HELP) {
		showUsage()
		os.Exit(0)
	}

	switch len(opts) {
	case 0:
		showUsage()
	case 1:
		process(opts[0])
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// intSignalHandler handles SIGINT signal and stops the program
func intSignalHandler() {
	piper.Close()
	os.Exit(0)
}

// termSignalHandler handles SIGTERM signal and stops the program
func termSignalHandler() {
	piper.Close()
	os.Exit(0)
}

// usr1SignalHandler handles SIGUSR1 signal
func usr1SignalHandler() {
	piper.Reopen()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// timeNow returns string of current datetime
func timeNow() string {
	return time.Now().Format(TS_PIPER_FORMAT)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// registerSignalHandlers registers handlers for signals
func registerSignalHandlers() {
	signal.Handlers{
		signal.TERM: termSignalHandler,
		signal.INT:  intSignalHandler,
		signal.USR1: usr1SignalHandler,
	}.TrackAsync()
}

// setupPiperOutput sets up logging file parameters
func setupPiperOutput(logFile string) {
	err := piper.Set(logFile, 0644)

	if err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// rotateOutput rotates file
func rotateOutput() error {
	fmt.Printf("Rotate file %s\n", piper.Path())
	piper.Close()

	newPath := fmt.Sprintf("%s.%v", piper.Path(), time.Now().UnixNano())

	if err := os.Rename(piper.Path(), newPath); err != nil {
		return err
	}

	return piper.Reopen()
}

// getSizeLimit provides size in bytes for size-limit log rotation
func getSizeLimit() uint64 {
	return fmtutil.ParseSize(options.GetS(OPT_SIZELIMIT))
}

// checkSizeLimit returns true if is it time to rotate file by size
func checkSizeLimit(line string, sizeLimit uint64) bool {
	return (sizeLimit != 0) && (piper.Size()+uint64(len(line)) > sizeLimit)
}

// prependTimestamp adds timestamp before line
func prependTimestamp(line string) string {
	return fmt.Sprintf("[%s] %s", timeNow(), line)
}

// writeLog writes log entry to file and stdout
func writeLog(line string) {
	fmt.Println(line)
	piper.Write([]byte(line))
	piper.Write([]byte{'\n'})
}

// runPiper starts reading stream from stdin and writing to log
func runPiper() error {
	scanner := bufio.NewScanner(os.Stdin)

	sizeLimit := getSizeLimit()

	fmt.Printf("DEBUG: size-limit=%v timestamp=%t\n", sizeLimit, options.GetB(OPT_TIMESTAMP))

	for scanner.Scan() {
		line := scanner.Text()

		if options.GetB(OPT_TIMESTAMP) {
			line = prependTimestamp(line)
		}

		if checkSizeLimit(line, sizeLimit) {
			if err := rotateOutput(); err != nil {
				return err
			}
		}

		writeLog(line)
	}

	return scanner.Err()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Process arguments
func process(logFile string) {
	setupPiperOutput(logFile)
	registerSignalHandlers()

	if err := runPiper(); err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// shutdown finishes program with return-code 1
func shutdown() {
	os.Exit(1)
}

// printErrorMessageAndExit finishes program with an error message
func printErrorMessageAndExit(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	shutdown()
}

// Show usage info
func showUsage() {
	info := usage.NewInfo(APP, "path")

	info.AddOption(OPT_SIZELIMIT, "Max file size", "size")
	info.AddOption(OPT_TIMESTAMP, "Prepend timestamp to every entry")
	info.AddOption(OPT_NO_COLOR, "Disable colored output")
	info.AddOption(OPT_VERSION, "Show information about version")
	info.AddOption(OPT_HELP, "Show this help message")

	info.AddExample("/var/log/program.log", "Read info from stdin and write to logging file")
	info.AddExample("/var/log/program.log -S 1024kb", "Rotate log every 1024 kilobytes")

	info.Render()
}

// Show info about version and license
func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2007,
		Owner:   "Gleb E Goncharov",
		License: "MIT",
	}

	about.Render()
}
