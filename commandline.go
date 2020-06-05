package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/commandline"
	"github.com/sbreitf1/go-console/input"
)

var (
	colorPromptRRI             = "\033[1;34m"
	colorPromptUser            = "\033[1;32m"
	colorPromptHost            = "\033[1;32m"
	colorPromptDry             = "\033[1;31m"
	colorPrintDry              = "\033[0;29m"
	colorSendRaw               = "\033[0;94m"
	colorReceiveRaw            = "\033[0;96m"
	colorSuccessResponse       = "\033[0;29m"
	colorErrorResponseMessage  = "\033[0;91m"
	colorTechnicalErrorMessage = "\033[1;91m"
	colorEnd                   = "\033[0m"
)

func disableColors() {
	colorPromptRRI = ""
	colorPromptUser = ""
	colorPromptHost = ""
	colorPromptDry = ""
	colorPrintDry = ""
	colorSendRaw = ""
	colorReceiveRaw = ""
	colorSuccessResponse = ""
	colorErrorResponseMessage = ""
	colorTechnicalErrorMessage = ""
	colorEnd = ""
}

func runCLE(client *rri.Client) error {
	if !console.SupportsColors() {
		disableColors()
	}

	cle := prepareCLE(client)
	err := cle.Run()
	if err != nil {
		if commandline.IsErrCtrlC(err) {
			console.Println()
			return nil
		}
		return err
	}
	return nil
}

func prepareCLE(client *rri.Client) *commandline.Environment {
	cle := commandline.NewEnvironment()
	cle.Prompt = func() string {
		var prefix, user, host, suffix string
		prefix = fmt.Sprintf("%sRRI%s", colorPromptRRI, colorEnd)
		if client.IsLoggedIn() {
			user = fmt.Sprintf("%s%s%s@", colorPromptUser, client.CurrentUser(), colorEnd)
		}
		host = fmt.Sprintf("%s%s%s", colorPromptHost, client.RemoteAddress(), colorEnd)
		if client.Processor != nil {
			suffix = fmt.Sprintf("%sDRY%s", colorPromptDry, colorEnd)
		}
		return fmt.Sprintf("%s{%s%s}%s", prefix, user, host, suffix)
	}
	cle.ErrorHandler = func(cmd string, args []string, err error) error {
		console.Printlnf("%sERROR: %s%s", colorTechnicalErrorMessage, err.Error(), colorEnd)
		return nil
	}
	cle.RegisterCommand(commandline.NewExitCommand("exit"))
	cle.RegisterCommand(commandline.NewCustomCommand("help", nil, cmdHelp))

	cle.RegisterCommand(commandline.NewCustomCommand("login", nil, newCommandWrapper(client, cmdLogin)))
	cle.RegisterCommand(commandline.NewCustomCommand("logout", nil, newCommandWrapper(client, cmdLogout)))

	cle.RegisterCommand(commandline.NewCustomCommand("create",
		commandline.NewFixedArgCompletion(commandline.NewOneOfArgCompletion("domain")),
		newCommandMultiplexer(client, []subCommand{
			subCommand{"domain", cmdCreateDomain},
		})))

	cle.RegisterCommand(commandline.NewCustomCommand("check",
		commandline.NewFixedArgCompletion(commandline.NewOneOfArgCompletion("domain")),
		newCommandMultiplexer(client, []subCommand{
			subCommand{"domain", newDomainQueryCommand(rri.NewCheckDomainQuery)},
		})))

	cle.RegisterCommand(commandline.NewCustomCommand("info",
		commandline.NewFixedArgCompletion(commandline.NewOneOfArgCompletion("domain")),
		newCommandMultiplexer(client, []subCommand{
			subCommand{"domain", newDomainQueryCommand(rri.NewInfoDomainQuery)},
		})))

	cle.RegisterCommand(commandline.NewCustomCommand("update",
		commandline.NewFixedArgCompletion(commandline.NewOneOfArgCompletion("domain")),
		newCommandMultiplexer(client, []subCommand{
			subCommand{"domain", cmdUpdateDomain},
		})))

	cle.RegisterCommand(commandline.NewCustomCommand("delete",
		commandline.NewFixedArgCompletion(commandline.NewOneOfArgCompletion("domain")),
		newCommandMultiplexer(client, []subCommand{
			subCommand{"domain", newDomainQueryCommand(rri.NewDeleteDomainQuery)},
		})))

	cle.RegisterCommand(commandline.NewCustomCommand("restore", nil, newCommandWrapper(client, newDomainQueryCommand(rri.NewRestoreDomainQuery))))
	//cle.RegisterCommand(commandline.NewCustomCommand("authinfo1", nil, newCommandWrapper(client, cmdCreateAuthInfo1)))
	//cle.RegisterCommand(commandline.NewCustomCommand("authinfo2", nil, newDomainCommandWrapper(client, rri.NewCreateAuthInfo2Query)))
	cle.RegisterCommand(commandline.NewCustomCommand("chprov", nil, newCommandWrapper(client, cmdChProv)))

	cle.RegisterCommand(commandline.NewCustomCommand("raw", nil, newCommandWrapper(client, cmdRaw)))
	cle.RegisterCommand(commandline.NewCustomCommand("file", commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(true)), newCommandWrapper(client, cmdFile)))

	cle.RegisterCommand(commandline.NewCustomCommand("xml", nil, newCommandWrapper(client, cmdXML)))
	cle.RegisterCommand(commandline.NewCustomCommand("verbose", nil, newCommandWrapper(client, cmdVerbose)))
	cle.RegisterCommand(commandline.NewCustomCommand("dry", nil, newCommandWrapper(client, cmdDry)))
	return cle
}

func cmdHelp(args []string) error {
	console.Println("Available commands:")
	console.Println("  exit                                -  exit application")
	console.Println("  help                                -  show this help")
	console.Println()
	console.Println("  login {user} {password}             -  log in to a RRI account")
	console.Println("  logout                              -  log out from the current RRI account")
	console.Println()
	//contact-create
	//contact-request
	//contact-check
	//contact-update
	//contact-info
	console.Println("  create domain {domain}              -  send a CREATE command for a new domain")
	console.Println("  check domain {domain}               -  send a CHECK command for a specific domain")
	console.Println("  info domain {domain}                -  send an INFO command for a specific domain")
	console.Println("  update domain {domain}              -  send an UPDATE command for a specific domain")
	//chholder
	console.Println("  delete domain {domain}              -  send a DELETE command for a specific domain")
	console.Println("  restore {domain}                    -  send a RESTORE command for a specific domain")
	console.Println("  create authinfo1 {domain} {secret}  -  send a CREATE-AUTHINFO1 command for a specific domain")
	console.Println("  create authinfo2 {domain}           -  send a CREATE-AUTHINFO2 command for a specific domain")
	//delete-authinfo1
	console.Println("  chprov {domain} {secret}            -  send a CHPROV command for a specific domain with auth info secret")
	//transit
	//
	//queue-read
	//queue-delete
	//
	//regacc-info
	console.Println()
	console.Println("  raw                                 -  enter a raw query and send it")
	console.Println("  file {path}                         -  process a query file as accepted by flag --file")
	console.Println()
	console.Println("  xml                                 -  toggle XML mode")
	console.Println("  verbose                             -  toggle verbose mode")
	console.Println("  dry                                 -  toggle dry mode to only print out raw queries")
	return nil
}

type clientCommand func(client *rri.Client, args []string) error

func newCommandWrapper(client *rri.Client, f clientCommand) commandline.ExecCommandHandler {
	return func(args []string) error {
		return f(client, args)
	}
}

type subCommand struct {
	Cmd     string
	Handler clientCommand
}

func newCommandMultiplexer(client *rri.Client, subCommands []subCommand) commandline.ExecCommandHandler {
	subMap := make(map[string]clientCommand)
	for _, c := range subCommands {
		subMap[strings.ToLower(c.Cmd)] = c.Handler
	}

	return func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing command type")
		}

		if handler, ok := subMap[strings.ToLower(args[0])]; ok {
			return handler(client, args[1:])
		}
		return fmt.Errorf("unknown command type")
	}
}

func newDomainQueryCommand(f func(domain string) *rri.Query) clientCommand {
	return newSingleArgQueryCommand("missing domain name", f)
}

func newSingleArgQueryCommand(missingMsg string, f func(arg string) *rri.Query) clientCommand {
	return func(client *rri.Client, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf(missingMsg)
		}

		_, err := processQuery(client, f(args[0]))
		return err
	}
}

func cmdLogin(client *rri.Client, args []string) error {
	var pass string
	if len(args) < 1 {
		return fmt.Errorf("missing RRI username and password")
	} else if len(args) < 2 {
		console.Println("Enter RRI password:")
		console.Print("> ")
		var err error
		pass, err = console.ReadPassword()
		if err != nil {
			return err
		}
	} else {
		pass = args[1]
	}

	if client.IsLoggedIn() {
		console.Println("Exiting current session")
		client.Logout()
	}

	return client.Login(args[0], pass)
}

func cmdLogout(client *rri.Client, args []string) error {
	return client.Logout()
}

func cmdCreateDomain(client *rri.Client, args []string) error {
	domainName, handles, nameServers, err := readDomainData(args, 0)
	if err != nil {
		return err
	}

	_, err = processQuery(client, rri.NewCreateDomainQuery(domainName, handles[0], handles[1], handles[2], nameServers...))
	return err
}

func cmdUpdateDomain(client *rri.Client, args []string) error {
	domainName, handles, nameServers, err := readDomainData(args, 0)
	if err != nil {
		return err
	}

	//TODO use old values for empty fields

	_, err = processQuery(client, rri.NewUpdateDomainQuery(domainName, handles[0], handles[1], handles[2], nameServers...))
	return err
}

func cmdCreateAuthInfo1(client *rri.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	_, err := processQuery(client, rri.NewCreateAuthInfo1Query(args[0], args[1]))
	return err
}

func cmdChProv(client *rri.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	domainName, handles, nameServers, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	_, err = processQuery(client, rri.NewChangeProviderQuery(domainName, args[1], handles[0], handles[1], handles[2], nameServers...))
	return err
}

func cmdRaw(client *rri.Client, args []string) error {
	raw, ok, err := input.Text("")
	if err != nil {
		return err
	}

	if ok {
		response, err := client.SendRaw(raw)
		if err != nil {
			return err
		}
		console.Println(response)
	}

	return nil
}

func cmdFile(client *rri.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing query file")
	}

	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	queries, err := rri.ParseQueries(string(data))
	if err != nil {
		return err
	}

	skipAuthQueries := client.IsLoggedIn()
	if skipAuthQueries {
		console.Println("Currently logged in. Auth queries will be skipped")
	}

	for _, query := range queries {
		if skipAuthQueries {
			if query.Action() == rri.ActionLogin || query.Action() == rri.ActionLogout {
				// skip authorization queries
				console.Println("Skip query", query) //TODO colored orange
				continue
			}
		}

		//TODO better visibility during dry mode

		console.Println("Exec query", query)
		success, err := processQuery(client, query)
		if err != nil {
			return err
		}
		if !success {
			break
		}
	}

	return nil
}

func cmdXML(client *rri.Client, args []string) error {
	client.XMLMode = !client.XMLMode
	if client.XMLMode {
		console.Println("XML mode on")
	} else {
		console.Println("XML mode off")
	}
	return nil
}

func cmdVerbose(client *rri.Client, args []string) error {
	if client.RawQueryPrinter == nil {
		client.RawQueryPrinter = rawQueryPrinter
		console.Println("Verbose mode on")
	} else {
		client.RawQueryPrinter = nil
		console.Println("Verbose mode off")
	}
	return nil
}

func cmdDry(client *rri.Client, args []string) error {
	if client.Processor == nil {
		client.Processor = dryProcessor
		console.Println("Dry mode on")
	} else {
		client.Processor = nil
		console.Println("Dry mode off")
	}
	return nil
}

func rawQueryPrinter(msg string, isOutgoing bool) {
	if isOutgoing {
		console.Printlnf("<-- %s%q%s", colorSendRaw, msg, colorEnd)
	} else {
		console.Printlnf("--> %s%q%s", colorReceiveRaw, msg, colorEnd)
	}
}

func dryProcessor(query *rri.Query) *rri.Query {
	console.Printlnf("%s%s%s", colorPrintDry, query.Export(), colorEnd)
	return nil
}

func processQuery(client *rri.Client, query *rri.Query) (bool, error) {
	response, err := client.SendQuery(query)
	if err != nil {
		return false, err
	}

	if response == nil {
		return true, nil
	}

	if response.IsSuccessful() {
		console.Print(colorSuccessResponse)
		for _, field := range response.Fields() {
			console.Printlnf("  %s: %s", field.Name, field.Value)
		}
		for _, entityName := range response.EntityNames() {
			console.Println()
			console.Printlnf("[%s]", entityName)
			entity := response.Entity(entityName)
			for _, field := range entity {
				console.Printlnf("  %s: %s", field.Name, field.Value)
			}
		}
		console.Print(colorEnd)
	} else {
		console.Printlnf("%sFailed: %s%s", colorErrorResponseMessage, response.ErrorMsg(), colorEnd)
	}
	return response.IsSuccessful(), nil
}

func readDomainData(args []string, dataOffset int) (domainName string, handles [][]string, nameServers []string, err error) {
	if len(args) < 1 {
		return "", nil, nil, fmt.Errorf("missing domain name")
	}

	domainName = args[0]
	if !strings.HasSuffix(strings.ToLower(domainName), ".de") {
		return "", nil, nil, fmt.Errorf("domain name must end with .de")
	}

	handleNames := []string{"Holder", "AbuseContact", "GeneralRequest"}
	handles = make([][]string, len(handleNames))
	for i := 0; i < len(handleNames); i++ {
		if len(args) >= (i + dataOffset + 1) {
			handles[i] = []string{args[i+dataOffset]}
		} else {
			console.Printf("%s Handle> ", handleNames[i])
			str, err := console.ReadLine()
			if err != nil {
				return "", nil, nil, err
			}
			handles[i] = []string{str}
		}
	}

	if len(args) >= (dataOffset + len(handleNames) + 1) {
		nameServers = args[dataOffset+len(handleNames)+1:]
	} else {
		for {
			console.Printf("NameServer> ")
			str, err := console.ReadLine()
			if err != nil {
				return "", nil, nil, err
			}
			if len(str) == 0 {
				break
			}
			nameServers = append(nameServers, str)
		}
	}

	return
}
