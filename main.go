package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/DENICeG/go-rriclient/internal/env"
	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/sbreitf1/go-console"
	"gopkg.in/alecthomas/kingpin.v2"
)

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
	argFail          = app.Flag("fail", "Exit with code 1 if RRI returns a failed result").Bool()
	argVerbose       = app.Flag("verbose", "Print all sent and received requests").Short('v').Bool()
	argInsecure      = app.Flag("insecure", "Disable SSL Certificate checks").Bool()
	argVersion       = app.Flag("version", "Display application version and exit").Bool()
	argDumpCLIConfig = app.Flag("dump-cli-config", "Print all configured colors and signs for testing").Bool()
)

type environment struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"pass" jcrypt:"aes"`
	Insecure bool   `json:"insecure"`
}

func (e environment) HasCredentials() bool {
	return len(e.User) > 0 && len(e.Password) > 0
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *argVersion {
		console.Println("Standalone RRI Client")
		if len(buildTime) == 0 || len(gitCommit) == 0 {
			console.Println("  no build information available")
		} else {
			console.Printlnf("  built at %s from commit %s", buildTime, gitCommit)
		}
		return
	}

	if *argDumpCLIConfig {
		console.Println("print colors and signs for testing:")
		printColorsAndSigns()
		return
	}

	if err := func() error {
		envReader, err := env.NewReader(".rri-client")
		if err != nil {
			return err
		}
		envReader.EnterEnvHandler = enterEnvironment
		envReader.GetEnvFileTitle = getEnvTitle

		if len(*argDeleteEnv) > 0 {
			if err := envReader.DeleteEnvironment(*argDeleteEnv); err != nil {
				return err
			}
			console.Printlnf("environment %q has been deleted", *argDeleteEnv)
			return nil
		}

		env, err := retrieveEnvironment(envReader)
		if err != nil {
			return err
		}

		if len(env.Address) == 0 {
			return fmt.Errorf("missing RRI server address")
		}

		client, err := rri.NewClient(env.Address, &rri.ClientConfig{Insecure: env.Insecure || *argInsecure})
		if err != nil {
			if !*argInsecure && strings.Contains(err.Error(), "x509") {
				// show help message for x509 related errors
				console.Println("HINT: try the '--insecure' flag if you have trouble with self signed certificates")
			}
			return err
		}
		defer client.Close()
		if *argVerbose {
			client.RawQueryPrinter = rawQueryPrinter
			client.InnerErrorPrinter = errorPrinter
		}

		if env.HasCredentials() {
			if err := client.Login(env.User, env.Password); err != nil {
				return err
			}
		}

		if len(*argFile) > 0 {
			content, err := ioutil.ReadFile(*argFile)
			if err != nil {
				return err
			}

			queries, err := parseQueries(string(content))
			if err != nil {
				return err
			}

			for _, query := range queries {
				console.Println("Exec query", query)
				response, err := client.SendQuery(query)
				if err != nil {
					return err
				}
				if response != nil && !response.IsSuccessful() {
					console.Printlnf("Query failed: %s", response.ErrorMsg())
					break
				}
			}

		} else {
			returnErrorOnFail = *argFail
			return runCLE(envReader.Dir(), client, *argCmd)
		}

		return nil
	}(); err != nil {
		console.Printlnf("FATAL: %s", err.Error())
		os.Exit(1)
	}
}

func retrieveEnvironment(envReader *env.Reader) (environment, error) {
	var addressFromCommandLine string
	if len(*argHost) > 0 {
		addressFromCommandLine = *argHost
		if !strings.Contains(addressFromCommandLine, ":") {
			// did the user forget to specify a port? use default port
			addressFromCommandLine += ":51131"
		}

	} else {
		if len(*argCmd) >= 1 && strings.Contains((*argCmd)[0], ":") {
			// consume first command part as address for backwards compatibility
			addressFromCommandLine = (*argCmd)[0]
			*argCmd = (*argCmd)[1:]
		}
	}

	var err error
	var env environment
	if len(*argEnvironment) > 0 {
		err = envReader.CreateOrReadEnvironment(*argEnvironment, &env)
	} else if len(addressFromCommandLine) == 0 {
		err = envReader.SelectEnvironment(&env)
	}
	if err != nil {
		return environment{}, err
	}

	if len(addressFromCommandLine) > 0 {
		env.Address = addressFromCommandLine
	}
	if len(*argUser) > 0 {
		env.User = *argUser
	}
	if len(*argPassword) > 0 {
		env.Password = *argPassword
	}

	if len(env.User) > 0 && len(env.Password) == 0 {
		// ask for missing user credentials
		var err error
		console.Printlnf("Please enter RRI password for user %q", env.User)
		console.Print("> ")
		env.Password, err = console.ReadPassword()
		if err != nil {
			return environment{}, err
		}
	}

	return env, nil
}

func enterEnvironment(envName string, env interface{}) error {
	e, ok := env.(*environment)
	if !ok {
		panic(fmt.Sprintf("environment has unexpected type %T", env))
	}

	var err error

	console.Print("Address (Host:Port)> ")
	e.Address, err = console.ReadLine()
	if err != nil {
		return err
	}
	if !strings.Contains(e.Address, ":") {
		// did the user forget to specify a port? use default port
		e.Address += ":51131"
	}

	console.Print("User> ")
	e.User, err = console.ReadLine()
	if err != nil {
		return err
	}

	console.Print("Password> ")
	e.Password, err = console.ReadPassword()
	if err != nil {
		return err
	}

	return nil
}

func getEnvTitle(envName, envFile string) string {
	data, err := ioutil.ReadFile(envFile)
	if err != nil {
		return envName
	}

	type envPreview struct {
		Address string `json:"address"`
		User    string `json:"user"`
	}
	var env envPreview
	if err := json.Unmarshal(data, &env); err != nil {
		return envName
	}

	var suffix string
	if len(env.User) > 0 {
		suffix = fmt.Sprintf(" (%s@%s)", env.User, env.Address)
	} else {
		suffix = fmt.Sprintf(" (%s)", env.Address)
	}
	return fmt.Sprintf("%s%s", envName, suffix)
}

// parseQueries parses multiple queries separated by a =-= line from a string.
func parseQueries(str string) ([]*rri.Query, error) {
	lines := strings.Split(str, "\n")

	// each string in queryStrings contains a single, unparsed query
	queryStrings := make([]string, 0)
	appendQueryString := func(str string) {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			queryStrings = append(queryStrings, str)
		}
	}

	// separate at lines beginning with =-=
	var sb strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, "=-=") {
			appendQueryString(sb.String())
			sb.Reset()
		} else {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	if sb.Len() > 0 {
		appendQueryString(sb.String())
	}

	queries := make([]*rri.Query, len(queryStrings))
	for i, queryString := range queryStrings {
		query, err := rri.ParseQueryKV(strings.TrimSpace(queryString))
		if err != nil {
			return nil, err
		}
		queries[i] = query
	}
	return queries, nil
}
