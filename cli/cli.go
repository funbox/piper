package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gongled/piper/logging"

	"pkg.re/essentialkaos/ek.v9/fmtc"
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
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VERSION  = "v:version"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VERSION:  {Type: options.BOOL, Alias: "ver"},
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

// registerSignalHandlers registers handlers for signals
func registerSignalHandlers() {
	signal.Handlers{
		signal.TERM: termSignalHandler,
		signal.INT:  intSignalHandler,
		signal.USR1: usr1SignalHandler,
	}.TrackAsync()
}

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

// setupPiperOutput sets up logging file parameters
func setupPiperOutput(logFile string) {
	err := piper.Set(logFile, 0644)

	if err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// runPiper starts reading stream from stdin and writing to log
func runPiper() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
		piper.Write([]byte(fmt.Sprintln(scanner.Text())))
	}

	if err := scanner.Err(); err != nil {
		printErrorMessageAndExit(err.Error())
	}
}

// Process arguments
func process(logFile string) {
	setupPiperOutput(logFile)
	registerSignalHandlers()
	runPiper()
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

	info.AddOption(OPT_NO_COLOR, "Disable colored output")
	info.AddOption(OPT_VERSION, "Show information about version")
	info.AddOption(OPT_HELP, "Show this help message")

	info.AddExample("/var/log/program.log", "Read info from the /dev/stdin and write to logging file")

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
