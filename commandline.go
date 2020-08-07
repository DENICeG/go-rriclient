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

	cleRRIClient *rri.Client
	histDomains  *domainHistory
	histHandles  *handleHistory
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
	cleRRIClient = client
	histDomains = &domainHistory{make([]string, 0)}
	histHandles = &handleHistory{make([]string, 0)}

	if !console.SupportsColors() {
		disableColors()
	}

	cle := prepareCLE()

	console.Println("Interactive RRI Command Line")
	console.Println("  type 'help' to see a list of available commands")
	console.Println("  use tab for auto-completion and arrow keys for history")

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

func prepareCLE() *commandline.Environment {
	cle := commandline.NewEnvironment()
	cle.Prompt = func() string {
		var prefix, user, host, suffix string
		prefix = fmt.Sprintf("%sRRI%s", colorPromptRRI, colorEnd)
		if cleRRIClient.IsLoggedIn() {
			user = fmt.Sprintf("%s%s%s@", colorPromptUser, cleRRIClient.CurrentUser(), colorEnd)
		}
		host = fmt.Sprintf("%s%s%s", colorPromptHost, cleRRIClient.RemoteAddress(), colorEnd)
		if cleRRIClient.Processor != nil {
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

	cle.RegisterCommand(commandline.NewCustomCommand("login", nil, cmdLogin))
	cle.RegisterCommand(commandline.NewCustomCommand("logout", nil, cmdLogout))

	registerDomainOrHandleSwitchCommand(cle, "create", cmdCreateDomain, nil)
	registerDomainOrHandleSwitchCommand(cle, "check", newDomainQueryCommand(rri.NewCheckDomainQuery), newHandleQueryCommand(rri.NewCheckHandleQuery))
	registerDomainOrHandleSwitchCommand(cle, "info", newDomainQueryCommand(rri.NewInfoDomainQuery), newHandleQueryCommand(rri.NewInfoHandleQuery))
	registerDomainOrHandleSwitchCommand(cle, "update", cmdUpdateDomain, nil)

	registerDomainCommand(cle, "delete", newDomainQueryCommand(rri.NewDeleteDomainQuery))
	registerDomainCommand(cle, "restore", newDomainQueryCommand(rri.NewRestoreDomainQuery))
	//TODO authinfo1
	//TODO authinfo2
	registerDomainCommand(cle, "chprov", cmdChProv)

	cle.RegisterCommand(commandline.NewCustomCommand("raw", nil, cmdRaw))
	cle.RegisterCommand(commandline.NewCustomCommand("file", commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(true)), cmdFile))

	cle.RegisterCommand(commandline.NewCustomCommand("xml", nil, cmdXML))
	cle.RegisterCommand(commandline.NewCustomCommand("verbose", nil, cmdVerbose))
	cle.RegisterCommand(commandline.NewCustomCommand("dry", nil, cmdDry))
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
	console.Println("  check handle {domain}               -  send a CHECK command for a specific handle")
	console.Println("  info handle {domain}                -  send an INFO command for a specific handle")
	//contact-update
	console.Println()
	console.Println("  create domain {domain}              -  send a CREATE command for a new domain")
	console.Println("  check domain {domain}               -  send a CHECK command for a specific domain")
	console.Println("  info domain {domain}                -  send an INFO command for a specific domain")
	console.Println("  update domain {domain}              -  send an UPDATE command for a specific domain")
	//chholder
	console.Println("  delete {domain}                     -  send a DELETE command for a specific domain")
	console.Println("  restore {domain}                    -  send a RESTORE command for a specific domain")
	//console.Println("  create authinfo1 {domain} {secret}  -  send a CREATE-AUTHINFO1 command for a specific domain")
	//console.Println("  create authinfo2 {domain}           -  send a CREATE-AUTHINFO2 command for a specific domain")
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

type history interface {
	Put(str string)
}

func addToListRemoveDouble(list []string, new ...string) []string {
	newList := make([]string, 0)
	for _, str := range list {
		// filter out all values equal to one in new (ignore case)
		found := false
		for _, newStr := range new {
			if strings.ToLower(str) == strings.ToLower(newStr) {
				found = true
				break
			}
		}
		if !found {
			newList = append(newList, str)
		}
	}
	return append(list, new...)
}

type domainHistory struct {
	list []string
}

func (h *domainHistory) Put(domain string) {
	h.list = addToListRemoveDouble(h.list, domain)
}

func (h *domainHistory) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	return commandline.PrepareCompletionOptions(h.list, false)
}

type handleHistory struct {
	list []string
}

func (h *handleHistory) Put(handle string) {
	h.list = addToListRemoveDouble(h.list, handle)
}

func (h *handleHistory) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	options := commandline.PrepareCompletionOptions(h.list, false)
	regAccID, err := cleRRIClient.CurrentRegAccID()
	if err == nil {
		return append(options, commandline.PrepareCompletionOptions([]string{fmt.Sprintf("DENIC-%d-", regAccID)}, true)...)
	}
	return options
}

func (h *handleHistory) GetHistoryEntry(index int) (string, bool) {
	if index >= len(h.list) {
		return "", false
	}
	return h.list[index], true
}

type domainOrHandleCompletion struct{}

func (domainOrHandleCompletion) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	if len(currentCommand) >= 2 && entryIndex == 2 {
		if currentCommand[1] == "domain" {
			return histDomains.GetCompletionOptions(currentCommand, entryIndex)
		}

		if currentCommand[1] == "handle" {
			return histHandles.GetCompletionOptions(currentCommand, entryIndex)
		}
	}
	return nil
}

func registerDomainOrHandleSwitchCommand(cle *commandline.Environment, name string, domainCommandHandler, handleCommandHandler commandline.ExecCommandHandler) {
	cle.RegisterCommand(commandline.NewCustomCommand(name,
		commandline.NewFixedArgCompletion(
			commandline.NewOneOfArgCompletion("domain", "handle"),
			domainOrHandleCompletion{},
		),
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing command type")
			}

			// explicit type is given
			if args[0] == "domain" {
				// remove type parameter
				return domainCommandHandler(args[1:])
			}
			if args[0] == "handle" {
				// remove type parameter
				return handleCommandHandler(args[1:])
			}

			// try to guess type from first parameter
			if isDomainName(args[0]) {
				return domainCommandHandler(args)
			}
			if isHandle(args[0]) {
				return handleCommandHandler(args)
			}

			return fmt.Errorf("unknown command type. expect 'domain' or 'handle'")
		}))
}

func isDomainName(str string) bool {
	return strings.HasSuffix(str, ".de")
}

func isHandle(str string) bool {
	return strings.HasPrefix(str, "DENIC-")
}

func registerDomainCommand(cle *commandline.Environment, name string, commandHandler commandline.ExecCommandHandler) {
	cle.RegisterCommand(commandline.NewCustomCommand(name, commandline.NewFixedArgCompletion(histDomains), commandHandler))
}

func newDomainQueryCommand(f func(domain string) *rri.Query) commandline.ExecCommandHandler {
	return newSingleArgQueryCommand("missing domain name", histDomains, f)
}

func newHandleQueryCommand(f func(domain string) *rri.Query) commandline.ExecCommandHandler {
	return newSingleArgQueryCommand("missing handle", histHandles, f)
}

func newSingleArgQueryCommand(missingMsg string, hist history, f func(arg string) *rri.Query) commandline.ExecCommandHandler {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf(missingMsg)
		}

		_, err := processQuery(f(args[0]))
		if hist != nil {
			hist.Put(args[0])
		}
		return err
	}
}

func cmdLogin(args []string) error {
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

	if cleRRIClient.IsLoggedIn() {
		console.Println("Exiting current session")
		cleRRIClient.Logout()
	}

	return cleRRIClient.Login(args[0], pass)
}

func cmdLogout(args []string) error {
	return cleRRIClient.Logout()
}

func cmdCreateDomain(args []string) error {
	domainName, handles, nameServers, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	_, err = processQuery(rri.NewCreateDomainQuery(domainName, handles[0], handles[1], handles[2], nameServers...))
	histDomains.Put(domainName)
	return err
}

func cmdUpdateDomain(args []string) error {
	domainName, handles, nameServers, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	//TODO use old domain values for empty fields -> only change explicitly entered data

	_, err = processQuery(rri.NewUpdateDomainQuery(domainName, handles[0], handles[1], handles[2], nameServers...))
	histDomains.Put(domainName)
	return err
}

func cmdCreateAuthInfo1(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	_, err := processQuery(rri.NewCreateAuthInfo1Query(args[0], args[1]))
	return err
}

func cmdChProv(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	domainName, handles, nameServers, err := readDomainData(args, 2)
	if err != nil {
		return err
	}

	_, err = processQuery(rri.NewChangeProviderQuery(domainName, args[1], handles[0], handles[1], handles[2], nameServers...))
	histDomains.Put(domainName)
	return err
}

func cmdRaw(args []string) error {
	raw, ok, err := input.Text("")
	if err != nil {
		return err
	}

	if ok {
		response, err := cleRRIClient.SendRaw(raw)
		if err != nil {
			return err
		}
		console.Println(response)
	}

	return nil
}

func cmdFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing query file")
	}

	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	queries, err := parseQueries(string(data))
	if err != nil {
		return err
	}

	skipAuthQueries := cleRRIClient.IsLoggedIn()
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
		success, err := processQuery(query)
		if err != nil {
			return err
		}
		if !success {
			break
		}
	}

	return nil
}

func cmdXML(args []string) error {
	cleRRIClient.XMLMode = !cleRRIClient.XMLMode
	if cleRRIClient.XMLMode {
		console.Println("XML mode on")
	} else {
		console.Println("XML mode off")
	}
	return nil
}

func cmdVerbose(args []string) error {
	if cleRRIClient.RawQueryPrinter == nil {
		cleRRIClient.RawQueryPrinter = rawQueryPrinter
		console.Println("Verbose mode on")
	} else {
		cleRRIClient.RawQueryPrinter = nil
		console.Println("Verbose mode off")
	}
	return nil
}

func cmdDry(args []string) error {
	if cleRRIClient.Processor == nil {
		cleRRIClient.Processor = dryProcessor
		console.Println("Dry mode on")
	} else {
		cleRRIClient.Processor = nil
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
	console.Printlnf("%s%s%s", colorPrintDry, query.EncodeKV(), colorEnd)
	return nil
}

func processQuery(query *rri.Query) (bool, error) {
	response, err := cleRRIClient.SendQuery(query)
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

	handleNames := []string{"Holder", "GeneralRequest", "AbuseContact"}
	handles = make([][]string, len(handleNames))
	for i := 0; i < len(handleNames); i++ {
		if len(args) >= (i + dataOffset + 1) {
			handles[i] = []string{args[i+dataOffset]}
		} else {
			console.Printf("%s Handle> ", handleNames[i])
			str, err := commandline.ReadLineWithHistory(histHandles)
			if err != nil {
				return "", nil, nil, err
			}
			histHandles.Put(str)
			handles[i] = []string{str}
		}
	}

	if len(args) >= (dataOffset + len(handleNames)) {
		nameServers = args[dataOffset+len(handleNames):]
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
