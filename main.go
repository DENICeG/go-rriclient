package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DENICeG/go-rriclient/internal/env"
	"github.com/DENICeG/go-rriclient/pkg/cli"
	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sbreitf1/go-console"
)

const (
	// this field is not updated automatically and needs to be set before every release!
	version = "1.11.0"
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
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	envReader, err := env.NewReader(".rri-client")
	if err != nil {
		logAndExit(err)
	}

	envReader.EnterEnvHandler = cli.EnterEnvironment
	envReader.GetEnvFileTitle = env.GetEnvTitle

	printVersion(argVersion, buildTime, gitCommit)
	printEnv(argListEnv, envReader)
	deleteEnv(argDeleteEnv, envReader)

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

	cliService := cli.New(client)

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

func deleteEnv(argDeleteEnv *string, envReader *env.Reader) {
	if len(*argDeleteEnv) == 0 {
		return
	}

	err := envReader.DeleteEnvironment(*argDeleteEnv)
	if err != nil {
		logAndExit(err)
	}
	console.Printlnf("environment %q has been deleted", *argDeleteEnv)
	os.Exit(0)
}

func printEnv(argListEnv *bool, envReader *env.Reader) {
	if !*argListEnv {
		return
	}

	environments, err := envReader.ListEnvironments() //nolint
	if err != nil {
		logAndExit(err)
	}

	for _, env := range environments {
		console.Printlnf("- %s", env)
	}

	os.Exit(0)
}

func printVersion(argVersion *bool, buildTime, gitCommit string) {
	if !*argVersion {
		return
	}

	console.Printlnf("Standalone RRI Client v%s", version)
	if len(buildTime) > 0 && len(gitCommit) > 0 {
		console.Printlnf("  built at %s from commit %s", buildTime, gitCommit)
	}

	os.Exit(0)
}

func logAndExit(err error) {
	console.Printlnf("FATAL: %s", err.Error())
	os.Exit(1)
}
