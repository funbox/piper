package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gongled/piper/logging"

	"pkg.re/essentialkaos/ek.v9/fmtc"
	"pkg.re/essentialkaos/ek.v9/fmtutil"
	"pkg.re/essentialkaos/ek.v9/timeutil"
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
	OPT_SIZELIMIT = "s:size"
	OPT_KEEPFILES = "k:keep"
	OPT_TIMELIMIT = "a:age"
	OPT_TIMESTAMP = "t:timestamp"
	OPT_NO_COLOR  = "nc:no-color"
	OPT_HELP      = "h:help"
	OPT_VERSION   = "v:version"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_SIZELIMIT: {Type: options.STRING, Bound: OPT_KEEPFILES},
	OPT_KEEPFILES: {Type: options.INT, Value: 0},
	OPT_TIMELIMIT: {Type: options.STRING, Bound: OPT_KEEPFILES},
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
			printErrorMessageAndExit(err.Error())
		}
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
	log.Close()
	os.Exit(0)
}

// termSignalHandler handles SIGTERM signal and stops the program
func termSignalHandler() {
	log.Close()
	os.Exit(0)
}

// usr1SignalHandler handles SIGUSR1 signal
func usr1SignalHandler() {
	log.RollOver()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// parseMaxTimeInterval parses duration to integer
func parseMaxTimeInterval(maxTimeInterval string) int64 {
	return timeutil.ParseDuration(maxTimeInterval)
}

// parseMaxFileSize parses file size from string to unsigned integer
func parseMaxFileSize(maxFileSize string) uint64 {
	return fmtutil.ParseSize(maxFileSize)
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
func setUpLogger(logFile string) {
	log.SetOutput(logFile)

	log.SetMaxTimeInterval(parseMaxTimeInterval(options.GetS(OPT_TIMELIMIT)))
	log.SetMaxFileSize(parseMaxFileSize(options.GetS(OPT_SIZELIMIT)))
	log.SetMaxBackupIndex(options.GetI(OPT_KEEPFILES))
	log.SetTimestampFlag(options.GetB(OPT_TIMESTAMP))

	err := log.Run()

	if err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// runPiper starts reading stream from stdin and writing to log
func runPiper() error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if err := log.WriteLog(scanner.Text()); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Process arguments
func process(logFile string) {
	setUpLogger(logFile)
	registerSignalHandlers()

	if err := runPiper(); err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// shutdown finishes program with return-code 1
func shutdown(exitCode int) {
	os.Exit(exitCode)
}

// printErrorMessageAndExit finishes program with an error message
func printErrorMessageAndExit(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	shutdown(1)
}

// Show usage info
func showUsage() {
	info := usage.NewInfo(APP, "path")

	info.AddOption(OPT_SIZELIMIT, "Max file size", "size")
	info.AddOption(OPT_KEEPFILES, "Number of files to keep", "number")
	info.AddOption(OPT_TIMELIMIT, "Interval of log rotation", "interval")
	info.AddOption(OPT_TIMESTAMP, "Prepend timestamp to every entry")
	info.AddOption(OPT_NO_COLOR, "Disable colored output")
	info.AddOption(OPT_VERSION, "Show information about version")
	info.AddOption(OPT_HELP, "Show this help message")

	info.AddExample("/var/log/program.log", "Read stdin and write to log")
	info.AddExample("-t /var/log/program.log", "Read stdin, prepend timestamp and write to log")
	info.AddExample("-s 5MB -k 10 /var/log/program.log", "Read stdin and rotate log every 5 megabytes and keep 10 files")
	info.AddExample("-a 10m -k 5 /var/log/program.log", "Read stdin and rotate log every 10 minute and keep 5 files")

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
