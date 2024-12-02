package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/DENICeG/go-rriclient/internal/env"
	"github.com/DENICeG/go-rriclient/pkg/cli"
	"github.com/DENICeG/go-rriclient/pkg/preset"
	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sbreitf1/go-console"
)

const (
	// this field is not updated automatically and needs to be set before every release!
	version = "1.11.0"
)

var (
	//go:embed examples
	embedFS embed.FS
)

func main() {
	var (
		gitCommit string
		buildTime string
	)

	var (
		app              = kingpin.New("rri-client", "Client application for RRI")
		argCmd           = app.Arg("command", "Command with arguments or RRI host like host:51131").Strings()
		argHost          = app.Flag("host", "A RRI host like host:51131").Short('h').String()
		argUser          = app.Flag("user", "RRI user to use for login").Short('u').String()
		argPassword      = app.Flag("pass", "RRI password to use for login. Will be asked for if only user is set").Short('p').String()
		argFile          = app.Flag("file", "Input file containing RRI requests separated by a '=-=' line").Short('f').String()
		argEnvironment   = app.Flag("env", "Named environment to use or create").Short('e').String()
		argDeleteEnv     = app.Flag("delete-env", "Delete an existing environment").String()
		argListEnv       = app.Flag("list-env", "List all environments").Bool()
		argFail          = app.Flag("fail", "Exit with code 1 if RRI returns a failed result").Bool()
		argVerbose       = app.Flag("verbose", "Print all sent and received requests").Short('v').Bool()
		argInsecure      = app.Flag("insecure", "Disable SSL Certificate checks").Bool()
		argVersion       = app.Flag("version", "Display application version and exit").Bool()
		argDumpCLIConfig = app.Flag("dump-cli-config", "Print all configured colors and signs for testing").Bool()
		argPreset        = app.Flag("preset", "Dynamically load, edit and execute a query from a preset").Short('P').Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	dirInfo, err := embedFS.ReadDir("examples")
	if err != nil {
		logAndExit(err)
	}

	presets, err := preset.Load(embedFS, dirInfo)
	if err != nil {
		logAndExit(err)
	}

	envReader, err := env.NewReader(".rri-client")
	if err != nil {
		logAndExit(err)
	}

	envReader.EnterEnvHandler = cli.EnterEnvironment
	envReader.GetEnvFileTitle = env.GetEnvTitle

	exit := cli.PrintVersion(argVersion, buildTime, gitCommit, version)
	shutdown(exit)

	exit, err = cli.PrintEnv(argListEnv, envReader)
	if err != nil {
		logAndExit(err)
	}

	shutdown(exit)

	exit, err = cli.DeleteEnv(argDeleteEnv, envReader)
	if err != nil {
		logAndExit(err)
	}

	shutdown(exit)

	env, err := cli.RetrieveEnvironment(envReader, argHost, argEnvironment, argUser, argPassword, argCmd)
	if err != nil {
		logAndExit(err)
	}

	if len(env.Address) == 0 {
		logAndExit(fmt.Errorf("missing RRI server address"))
	}

	client, err := rri.NewClient(env.Address, &rri.ClientConfig{Insecure: env.Insecure || *argInsecure})
	if err != nil {
		if !*argInsecure && strings.Contains(err.Error(), "x509") {
			// show help message for x509 related errors
			console.Println("HINT: try the '--insecure' flag if you have trouble with self signed certificates")
		}

		logAndExit(err)
	}

	defer client.Close()

	cliService := cli.New(client, *presets)

	if argPreset != nil && *argPreset == true {
		cliService.HandlePreset([]string{})
	}

	if *argDumpCLIConfig {
		console.Println("print colors and signs for testing:")
		cliService.PrintColorsAndSigns()
		return
	}

	if *argVerbose {
		client.RawQueryPrinter = cliService.RawQueryPrinter
		client.InnerErrorPrinter = cliService.ErrorPrinter
	}

	if env.HasCredentials() {
		err = client.Login(env.User, env.Password)
		if err != nil {
			logAndExit(err)
		}
	}

	if len(*argFile) > 0 {
		err = cliService.HandleFile([]string{*argFile})
		if err != nil {
			logAndExit(err)
		}
	}

	cliService.ReturnErrorOnFail = *argFail
	err = cliService.Run(envReader.Dir(), *argCmd)
	if err != nil {
		logAndExit(err)
	}
}

func shutdown(shutdown bool) {
	if shutdown {
		os.Exit(0)
	}
}

func logAndExit(err error) {
	console.Printlnf("FATAL: %s", err.Error())
	os.Exit(1)
}
