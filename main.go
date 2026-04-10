package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/do_api"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/logger"
)

func run() error {

	cmdHelp := flag.Bool("h", false, "Show the help message")
	cmdHelp2 := flag.Bool("help", false, "Show the help message")

	cmdDbg := flag.Bool("d", false, "Outputs log messages to the standard console")
	cmdDbg2 := flag.Bool("debug", false, "Outputs log messages to the standard console")
	flag.Parse()

	if *cmdHelp || *cmdHelp2 {
		usage()
		os.Exit(1)
	}

	if !((*cmdDbg) || (*cmdDbg2)) {
		// if no debug option was selected, discard all debug output
		logger.SetDebugOutput(io.Discard)
	} else {
		// default debug output to Stdout instead of Stderr
		logger.SetDebugOutput(os.Stdout)
	}
	configFile := ""
	if len(flag.Args()) == 0 {
		var err error
		// configFile, err = homedir.Dir()
		usr, err := user.Current()
		logger.CheckError(err)
		homeDir := usr.HomeDir
		if usr.HomeDir == "" {
			logger.Warning("Unable to determine the current user's home directory. Defaulting to current working directory. Consider specifying the config file path as a command argument or setting the HOME environment variable.")
			homeDir, err = os.Getwd()
			logger.CheckError(err)
		}
		configFile = homeDir
		configFile += "/.digitalocean-dynamic-ip.json"
	} else {
		configFile = flag.Args()[0]
	}
	logger.Debug("Using config file: %s", configFile)

	config.SetConfigFilePath(configFile)

	conf := config.Get()
	currentIPv4, currentIPv6 := CheckPublicIPs()
	if currentIPv4 == nil && currentIPv6 == nil {
		return fmt.Errorf("Current IP addresses are not valid, or both are disabled in the config. Check your configuration and internet connection.")
	}

	for _, domain := range conf.Domains {
		logger.Debug("%s: START", domain.Domain)
		do_api.UpdateRecords(domain, currentIPv4, currentIPv6)
		logger.Debug("%s: END", domain.Domain)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		logger.ErrorAndExit(err.Error())
	}
}

// usage prints the help message for command-line usage
func usage() {
	os.Stdout.WriteString(fmt.Sprintf("To use this program you can specify the following command options:\n"+
		"-h | -help\n\tShow this help message\n"+
		"-d | -debug\n\tPrint debug messages to standard output\n"+
		"[config_file]\n\tlocation of the configuration file\n\n"+
		"If the [config_file] parameter is not passed, then the default\n"+
		"config location of ~/.digitalocean-dynamic-ip.json will be used.\n\n"+
		"example usages:\n\t%[1]s -help\n"+
		"\t%[1]s\n"+
		"\t%[1]s %[2]s\n"+
		"\t%[1]s -debug %[2]s\n"+
		"",
		os.Args[0],
		"/path/to/my/config.json",
	))
}
