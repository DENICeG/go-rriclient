package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/commandline"
	"github.com/sbreitf1/go-console/input"
)

var (
	//TODO configure colors and signs from external file
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
	colorInnerError            = "\033[2;91m"
	colorEnd                   = "\033[0m"
	signSend                   = "-->"
	signReceive                = "<--"

	cleRRIClient *rri.Client
	histDomains  *domainHistory
	histHandles  *handleHistory

	returnErrorOnFail = false

	customCommands []customCommand
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
	colorInnerError = ""
	colorEnd = ""
}

func printColorsAndSigns() {
	printColor := func(name, colorStr string) {
		console.Printlnf("  %s%s%s", colorStr, name, colorEnd)
	}
	printSign := func(name, sign string) {
		console.Printlnf("  %s: %q", name, sign)
	}
	printColor("ColorDefault", colorEnd)
	printColor("ColorPromptRRI", colorPromptRRI)
	printColor("ColorPromptUser", colorPromptUser)
	printColor("ColorPromptHost", colorPromptHost)
	printColor("ColorPromptDry", colorPromptDry)
	printColor("ColorPrintDry", colorPrintDry)
	printColor("ColorSendRaw", colorSendRaw)
	printColor("ColorReceiveRaw", colorReceiveRaw)
	printColor("ColorSuccessResponse", colorSuccessResponse)
	printColor("ColorErrorResponseMessage", colorErrorResponseMessage)
	printColor("ColorTechnicalErrorMessage", colorTechnicalErrorMessage)
	printColor("ColorInnerError", colorInnerError)
	printSign("SignSend", signSend)
	printSign("SignReceive", signReceive)
}

func runCLE(confDir string, client *rri.Client, cmd []string) error {
	cleRRIClient = client
	histDomains = &domainHistory{make([]string, 0)}
	histHandles = &handleHistory{make([]string, 0)}

	if !console.SupportsColors() {
		disableColors()
	}

	var err error
	customCommands, err = readCustomCommands(filepath.Join(confDir, "custom-commands"))
	if err != nil {
		if !os.IsNotExist(err) {
			console.Println("Failed to import custom commands:", err.Error())
		}
	}
	cle := prepareCLE()

	if len(cmd) > 0 {
		// exec command that has been passed via command line and return result
		return cle.ExecCommand(cmd[0], cmd[1:])
	}

	console.Println("Interactive RRI Command Line")
	console.Println("  type 'help' to see a list of available commands")
	console.Println("  use tab for auto-completion and arrow keys for history")

	// start interactive command line loop
	if err := cle.Run(); err != nil {
		if commandline.IsErrCtrlC(err) {
			console.Println()
			return nil
		}
		return err
	}
	return nil
}

type customCommand struct {
	DisabledFor []string           `json:"disable-for-env"`
	Name        string             `json:"name"`
	Cmd         string             `json:"cmd"`
	Description string             `json:"description"`
	Action      string             `json:"action"`
	Args        []customCommandArg `json:"args"`
}

type customCommandArg struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Field string `json:"field"`
	Value string `json:"value"`
}

func (arg customCommandArg) IsInputParameter() bool {
	return strings.ToLower(arg.Type) == "domain"
}

func readCustomCommands(dir string) ([]customCommand, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	customCommands := make([]customCommand, 0)
	for _, f := range files {
		if !f.IsDir() {
			data, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				return nil, err
			}
			var cmd customCommand
			if err := json.Unmarshal(data, &cmd); err != nil {
				return nil, fmt.Errorf("could not import %q: %s", f.Name(), err.Error())
			}
			if len(cmd.Name) == 0 {
				cmd.Name = f.Name()
			}
			if len(cmd.Cmd) == 0 {
				cmd.Cmd = f.Name()
			}
			for i := range cmd.Args {
				if strings.ToLower(cmd.Args[i].Type) == "domain" && len(cmd.Args[i].Field) == 0 {
					cmd.Args[i].Field = "domain"
					cmd.Args[i].Name = "domain"
				}
			}
			customCommands = append(customCommands, cmd)
		}
	}
	return customCommands, nil
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

	registerSwitchCommand(cle, "create", cmdSwitches{
		Domain:    cmdCreateDomain,
		Handle:    cmdCreateHandle,
		AuthInfo1: cmdCreateAuthInfo1,
	})
	registerSwitchCommand(cle, "check", cmdSwitches{
		Domain: newDomainQueryCommand(rri.NewCheckDomainQuery),
		Handle: newHandleQueryCommand(rri.NewCheckHandleQuery),
	})
	registerSwitchCommand(cle, "info", cmdSwitches{
		Domain: newDomainQueryCommand(rri.NewInfoDomainQuery),
		Handle: newHandleQueryCommand(rri.NewInfoHandleQuery),
	})
	registerSwitchCommand(cle, "update", cmdSwitches{
		Domain: cmdUpdateDomain,
	})

	registerDomainCommand(cle, "delete", newDomainQueryCommand(rri.NewDeleteDomainQuery))
	registerDomainCommand(cle, "restore", newDomainQueryCommand(rri.NewRestoreDomainQuery))
	registerDomainCommand(cle, "transit", cmdTransit, commandline.NewOneOfArgCompletion("disconnect", "connect"))
	registerDomainCommand(cle, "chholder", cmdChangeHolder)
	registerDomainCommand(cle, "chprov", cmdChangeProvider)

	// register custom commands
	for _, cmd := range customCommands {
		registerCustomCommand(cle, cmd)
	}

	cle.RegisterCommand(commandline.NewCustomCommand("raw", nil, cmdRaw))
	cle.RegisterCommand(commandline.NewCustomCommand("file", commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(true)), cmdFile))

	cle.RegisterCommand(commandline.NewCustomCommand("xml", nil, cmdXML))
	cle.RegisterCommand(commandline.NewCustomCommand("verbose", nil, cmdVerbose))
	cle.RegisterCommand(commandline.NewCustomCommand("dry", nil, cmdDry))
	return cle
}

func cmdHelp(args []string) error {
	type customCmd struct {
		Cmd, Args []string
		Desc      string
	}
	commands := []customCmd{
		{[]string{"exit"}, nil, "exit application"},
		{[]string{"help"}, nil, "show this help"},
		{},
		{[]string{"login"}, []string{"user", "password"}, "log in to a RRI account"},
		{[]string{"logout"}, nil, "log out from the current RRI account"},
		{},
		{[]string{"create", "handle"}, []string{"domain"}, "send a CREATE command for a specific handle"},
		{[]string{"check", "handle"}, []string{"domain"}, "send a CHECK command for a specific handle"},
		{[]string{"info", "handle"}, []string{"domain"}, "send an INFO command for a specific handle"},
		//TODO contact-update
		{},
		{[]string{"create", "domain"}, []string{"domain"}, "send a CREATE command for a new domain"},
		{[]string{"check", "domain"}, []string{"domain"}, "send a CHECK command for a specific domain"},
		{[]string{"info", "domain"}, []string{"domain"}, "send an INFO command for a specific domain"},
		{[]string{"update", "domain"}, []string{"domain"}, "send an UPDATE command for a specific domain"},
		{[]string{"chholder", "domain"}, []string{"domain"}, "send an CHHOLDER command for a specific domain"},
		{},
		{[]string{"delete"}, []string{"domain"}, "send a DELETE command for a specific domain"},
		{[]string{"restore"}, []string{"domain"}, "send a RESTORE command for a specific domain"},
		{[]string{"transit"}, []string{"domain"}, "send a TRANSIT command for a specific domain"},
		{[]string{"create", "authinfo1"}, []string{"domain", "secret", "expire"}, "send a CREATE-AUTHINFO1 command for a specific domain"},
		//TODO create-authinfo2
		//TODO delete-authinfo1
		{[]string{"chprov"}, []string{"domain", "secret"}, "send a CHPROV command for a specific domain"},
		//TODO verify
		// -
		//TODO queue-read
		//TODO queue-delete
		//TODO regacc-info
		{},
		{[]string{"raw"}, nil, "enter a raw query and send it"},
		{[]string{"file"}, []string{"path"}, "process a query file as accepted by flag --file"},
		{},
		{[]string{"xml"}, nil, "toggle XML mode"},
		{[]string{"verbose"}, nil, "toggle verbose mode"},
		{[]string{"dry"}, nil, "toggle dry mode to only print out raw queries"},
	}

	if len(customCommands) > 0 {
		head := commands[:21]
		tail := make([]customCmd, 6)
		copy(tail, commands[21:])
		commands = head
		for _, cmd := range customCommands {
			args := make([]string, 0)
			for _, arg := range cmd.Args {
				if arg.IsInputParameter() {
					args = append(args, arg.Name)
				}
			}
			commands = append(commands, customCmd{[]string{cmd.Cmd}, args, cmd.Description})
		}
		commands = append(commands, customCmd{})
		commands = append(commands, tail...)
	}

	cmdSumStrings := make([]string, len(commands))
	argsSumStrings := make([]string, len(commands))

	console.Println("Available commands:")

	// measure command info texts to find the best fitting display variant for the current tty
	maxCmdSumLen := 0
	maxCmdLen := 0
	maxArgsSumLen := 0
	maxArgsLen := 0
	maxDescLen := 0
	for i, c := range commands {
		for _, a := range c.Cmd {
			if len(cmdSumStrings[i]) == 0 {
				cmdSumStrings[i] += a
			} else {
				cmdSumStrings[i] += " " + a
			}
			if len([]rune(a)) > maxCmdLen {
				maxCmdLen = len([]rune(a))
			}
		}
		if len([]rune(cmdSumStrings[i])) > maxCmdSumLen {
			maxCmdSumLen = len([]rune(cmdSumStrings[i]))
		}
		for _, a := range c.Args {
			if len(argsSumStrings[i]) == 0 {
				argsSumStrings[i] += "{" + a + "}"
			} else {
				argsSumStrings[i] += " {" + a + "}"
			}
			if len([]rune(a)) > maxArgsLen {
				maxArgsLen = len([]rune(a))
			}
		}
		if len([]rune(argsSumStrings[i])) > maxArgsSumLen {
			maxArgsSumLen = len([]rune(argsSumStrings[i]))
		}
		if len([]rune(c.Desc)) > maxDescLen {
			maxDescLen = len([]rune(c.Desc))
		}
	}

	// sum of spacers and placeholders in "  NAME ARGS_SUM  -  DESC"
	maxLineLength := 2 + maxCmdSumLen + 1 + maxArgsSumLen + 2 + 1 + 2 + maxDescLen
	w, _, err := console.GetSize()
	if w <= 0 || maxLineLength <= w || err != nil {
		// nice table for wide terminals
		for i, c := range commands {
			if c.Cmd == nil {
				console.Println()
			} else {
				leftLen := len([]rune(cmdSumStrings[i])) + 1 + len([]rune(argsSumStrings[i]))
				console.Printlnf("  %s %s%s  -  %s", cmdSumStrings[i], argsSumStrings[i], strings.Repeat(" ", maxCmdSumLen+1+maxArgsSumLen-leftLen), c.Desc)
			}
		}

	} else {
		if maxCmdSumLen+1+maxArgsSumLen <= w {
			// print narrow version with two separate lines for command+args and description
			for i, c := range commands {
				if c.Cmd == nil {
					console.Println()
				} else {
					if len(c.Args) > 0 {
						console.Printlnf("%s %s", cmdSumStrings[i], argsSumStrings[i])
					} else {
						console.Printlnf("%s", cmdSumStrings[i])
					}
					if len([]rune(c.Desc)) <= w-2 {
						console.Printlnf("  %s", c.Desc)

					} else {
						// break words of description to multiple lines with indentation
						parts := strings.Split(c.Desc, " ")
						lines := make([]string, 1)
						for _, p := range parts {
							l := lines[len(lines)-1]
							if len(l) > 0 {
								l += " " + p
							} else {
								l += p
							}
							if len([]rune(l)) < w-2 || len(lines[len(lines)-1]) == 0 {
								lines[len(lines)-1] = l
							} else {
								lines = append(lines, p)
							}
						}
						for _, l := range lines {
							console.Printlnf("  %s", l)
						}
					}
				}
			}

		} else {
			// fallback variant for very narrow terminals
			for i, c := range commands {
				if c.Cmd == nil {
					console.Println()
				} else {
					if len(c.Args) > 0 {
						console.Printlnf("-> %s %s - %s", cmdSumStrings[i], argsSumStrings[i], c.Desc)
					} else {
						console.Printlnf("-> %s - %s", cmdSumStrings[i], c.Desc)
					}
				}
			}
		}
	}
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
	return append(new, newList...)
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
		if currentCommand[1] == "domain" || currentCommand[1] == "authinfo1" {
			return histDomains.GetCompletionOptions(currentCommand, entryIndex)
		}

		if currentCommand[1] == "handle" {
			return histHandles.GetCompletionOptions(currentCommand, entryIndex)
		}
	}
	return nil
}

type cmdSwitches struct {
	Domain    commandline.ExecCommandHandler
	Handle    commandline.ExecCommandHandler
	AuthInfo1 commandline.ExecCommandHandler
}

func registerSwitchCommand(cle *commandline.Environment, name string, switches cmdSwitches) {
	// assemble arg completion
	types := make([]string, 0)
	if switches.Domain != nil {
		types = append(types, "domain")
	}
	if switches.Handle != nil {
		types = append(types, "handle")
	}
	if switches.AuthInfo1 != nil {
		types = append(types, "authinfo1")
	}

	cle.RegisterCommand(commandline.NewCustomCommand(name,
		commandline.NewFixedArgCompletion(
			commandline.NewOneOfArgCompletion(types...),
			domainOrHandleCompletion{},
		),
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing command type")
			}

			// explicit type is given
			if args[0] == "domain" && switches.Domain != nil {
				// remove type parameter
				return switches.Domain(args[1:])
			}
			if args[0] == "handle" && switches.Handle != nil {
				// remove type parameter
				return switches.Handle(args[1:])
			}
			if args[0] == "authinfo1" && switches.AuthInfo1 != nil {
				// remove type parameter
				return switches.AuthInfo1(args[1:])
			}

			// try to guess type from first parameter
			if isDomainName(args[0]) {
				if switches.Domain != nil {
					return switches.Domain(args)
				}
				if switches.AuthInfo1 != nil {
					return switches.AuthInfo1(args)
				}
			}
			if isHandle(args[0]) && switches.Handle != nil {
				return switches.Handle(args)
			}

			// assemble expect string
			str := ""
			for i, t := range types {
				switch {
				case i == 0:
					str += "'" + t + "'"
				case i == len(types)-1:
					str += " or '" + t + "'"
				default:
					str += ", '" + t + "'"
				}
			}
			return fmt.Errorf("unknown command type. expect %s", str)
		}))
}

func isDomainName(str string) bool {
	return strings.HasSuffix(str, ".de")
}

func isHandle(str string) bool {
	return strings.HasPrefix(str, "DENIC-")
}

func registerDomainCommand(cle *commandline.Environment, name string, commandHandler commandline.ExecCommandHandler, additionalCompletionHandlers ...commandline.ArgCompletion) {
	argCompletionHandlers := []commandline.ArgCompletion{histDomains}
	argCompletionHandlers = append(argCompletionHandlers, additionalCompletionHandlers...)
	cle.RegisterCommand(commandline.NewCustomCommand(name, commandline.NewFixedArgCompletion(argCompletionHandlers...), commandHandler))
}

func newDomainQueryCommand(f func(domain string) *rri.Query) commandline.ExecCommandHandler {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing domain name")
		}

		_, err := processQuery(f(args[0]))
		if histDomains != nil {
			histDomains.Put(args[0])
		}
		return err
	}
}

func newHandleQueryCommand(f func(handle rri.DenicHandle) *rri.Query) commandline.ExecCommandHandler {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing handle")
		}

		handle, err := rri.ParseDenicHandle(args[0])
		if err != nil {
			return err
		}

		_, err = processQuery(f(handle))
		if histHandles != nil {
			histHandles.Put(args[0])
		}
		return err
	}
}

func registerCustomCommand(cle *commandline.Environment, cmd customCommand) {
	clArgs := make([]commandline.ArgCompletion, 0)
	for _, arg := range cmd.Args {
		switch strings.ToLower(arg.Type) {
		case "domain":
			clArgs = append(clArgs, histDomains)
		}
	}
	cle.RegisterCommand(commandline.NewCustomCommand(cmd.Cmd, commandline.NewFixedArgCompletion(clArgs...), func(args []string) error {
		fields := rri.NewQueryFieldList()
		argIndex := 0
		for _, arg := range cmd.Args {
			var inValue string
			if arg.IsInputParameter() {
				if len(args) <= argIndex {
					return fmt.Errorf("missing argument '%s'", arg.Name)
				}
				inValue = args[argIndex]
				argIndex++
			}

			switch strings.ToLower(arg.Type) {
			case "domain":
				fields.Add(rri.QueryFieldName(arg.Field), inValue)
			case "const":
				fields.Add(rri.QueryFieldName(arg.Field), arg.Value)
			}
		}
		//TODO fill history
		query := rri.NewQuery(rri.LatestVersion, rri.QueryAction(cmd.Action), fields)
		_, err := processQuery(query)
		return err
	}))
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

func cmdCreateHandle(args []string) error {
	handle, contactData, err := readContactData(args, 0)
	if err != nil {
		return err
	}

	_, err = processQuery(rri.NewCreateContactQuery(handle, contactData))
	histHandles.Put(handle.String())
	return err
}

func cmdCreateDomain(args []string) error {
	domainName, domainData, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	_, err = processQuery(rri.NewCreateDomainQuery(domainName, domainData))
	histDomains.Put(domainName)
	return err
}

func cmdUpdateDomain(args []string) error {
	domainName, domainData, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	//TODO use old domain values for empty fields -> only change explicitly entered data

	_, err = processQuery(rri.NewUpdateDomainQuery(domainName, domainData))
	histDomains.Put(domainName)
	return err
}

func cmdChangeHolder(args []string) error {
	domainName, domainData, err := readDomainData(args, 1)
	if err != nil {
		return err
	}

	//TODO use old domain values for empty fields -> only change explicitly entered data

	_, err = processQuery(rri.NewChangeHolderQuery(domainName, domainData))
	histDomains.Put(domainName)
	return err
}

func cmdTransit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	domainName := args[0]
	var disconnect bool
	if len(args) > 1 {
		switch strings.ToLower(args[1]) {
		case "disconnect":
			disconnect = true
		case "connect":
			disconnect = false
		default:
			return fmt.Errorf("invalid mode '%s'. expecting 'disconnect' or 'connect'", args[1])
		}
	}
	_, err := processQuery(rri.NewTransitDomainQuery(domainName, disconnect))
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
	var expire time.Time
	if len(args) >= 3 {
		var err error
		expire, err = time.ParseInLocation("2006-01-02", args[2], time.Local)
		if err != nil {
			expire, err = time.ParseInLocation("20060102", args[2], time.Local)
			if err != nil {
				return fmt.Errorf("expiration date must be in format yyyy-mm-dd or yyyymmdd")
			}
		}

	} else {
		expire = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+7, 0, 0, 0, 0, time.Local)
		console.Println("using default expiration of 1 week")
	}

	_, err := processQuery(rri.NewCreateAuthInfo1Query(args[0], args[1], expire))
	return err
}

func cmdChangeProvider(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}
	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	domainName, domainData, err := readDomainData(args, 2)
	if err != nil {
		return err
	}

	_, err = processQuery(rri.NewChangeProviderQuery(domainName, args[1], domainData))
	histDomains.Put(domainName)
	return err
}

func cmdRaw(args []string) error {
	var rawCommand string
	if len(args) > 0 {
		rawCommand = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(args[0], "\\n", "\n"), "\\r", "\r"), "\\\\", "\\")

	} else {
		raw, ok, err := input.Text("")
		if err != nil {
			return err
		}
		if ok {
			rawCommand = raw
		}
	}

	if len(rawCommand) > 0 {
		response, err := cleRRIClient.SendRaw(rawCommand)
		if err != nil {
			return err
		}
		console.Println(response)

		if returnErrorOnFail {
			responseObj, err := rri.ParseResponse(response)
			if err != nil {
				return fmt.Errorf("RRI server returned an invalid response")
			}
			if !responseObj.IsSuccessful() {
				return fmt.Errorf("RRI returned result 'failed'")
			}
		}
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

	// find queries to execute first
	executedQueryCount := 0
	hasLoginLogoutQueries := false
	for _, query := range queries {
		if query.Action() == rri.ActionLogin || query.Action() == rri.ActionLogout {
			hasLoginLogoutQueries = true
			if !skipAuthQueries {
				executedQueryCount++
			}
		} else {
			executedQueryCount++
		}
	}

	if skipAuthQueries && hasLoginLogoutQueries {
		//TODO colored orange
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

		//TODO colored in send color
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
		cleRRIClient.InnerErrorPrinter = errorPrinter
		console.Println("Verbose mode on")
	} else {
		cleRRIClient.RawQueryPrinter = nil
		cleRRIClient.InnerErrorPrinter = nil
		console.Println("Verbose mode off")
	}
	return nil
}

func rawQueryPrinter(msg string, isOutgoing bool) {
	if isOutgoing {
		console.Printlnf("%s %s%q%s", signSend, colorSendRaw, rri.CensorRawMessage(msg), colorEnd)
	} else {
		console.Printlnf("%s %s%q%s", signReceive, colorReceiveRaw, rri.CensorRawMessage(msg), colorEnd)
	}
}

func errorPrinter(err error) {
	console.Printlnf("%sERR: %s%s", colorInnerError, err.Error(), colorEnd)
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

	var colorStr string
	if response.IsSuccessful() {
		colorStr = colorSuccessResponse
	} else {
		colorStr = colorErrorResponseMessage
	}

	console.Print(colorStr)
	for _, field := range response.Fields() {
		console.Printlnf("  %s: %s", field.Name, field.Value)
	}
	for _, entity := range response.Entities() {
		console.Println()
		console.Printlnf("[%s]", entity.Name())
		for _, field := range entity.Fields() {
			console.Printlnf("  %s: %s", field.Name, field.Value)
		}
	}
	console.Print(colorEnd)

	if returnErrorOnFail && !response.IsSuccessful() {
		return false, fmt.Errorf("RRI returned result 'failed'")
	}
	return response.IsSuccessful(), nil
}

func readDomainData(args []string, dataOffset int) (string, rri.DomainData, error) {
	if len(args) < 1 {
		return "", rri.DomainData{}, fmt.Errorf("missing domain name")
	}

	domainName := args[0]
	if !strings.HasSuffix(strings.ToLower(domainName), ".de") {
		return "", rri.DomainData{}, fmt.Errorf("domain name must end with .de")
	}

	handleNames := []string{"Holder", "GeneralRequest", "AbuseContact"}
	handles := make([]rri.DenicHandle, len(handleNames))
	for i := 0; i < len(handleNames); i++ {
		if len(args) >= (i + dataOffset + 1) {
			handle, err := rri.ParseDenicHandle(args[i+dataOffset])
			if err != nil {
				return "", rri.DomainData{}, fmt.Errorf("%q: %s", args[i+dataOffset], err.Error())
			}
			handles[i] = rri.DenicHandle(handle)
		} else {
			console.Printf("%s Handle> ", handleNames[i])
			str, err := commandline.ReadLineWithHistory(histHandles)
			if err != nil {
				return "", rri.DomainData{}, err
			}
			handle, err := rri.ParseDenicHandle(str)
			if err != nil {
				return "", rri.DomainData{}, fmt.Errorf("%q: %s", str, err.Error())
			}
			histHandles.Put(str)
			handles[i] = rri.DenicHandle(handle)
		}
	}

	var nameServers []string
	if len(args) >= (dataOffset + len(handleNames)) {
		nameServers = args[dataOffset+len(handleNames):]
	} else {
		for {
			console.Printf("NameServer> ")
			str, err := console.ReadLine()
			if err != nil {
				return "", rri.DomainData{}, err
			}
			if len(str) == 0 {
				break
			}
			nameServers = append(nameServers, str)
		}
	}

	return domainName, rri.DomainData{
		HolderHandles:         []rri.DenicHandle{handles[0]},
		GeneralRequestHandles: []rri.DenicHandle{handles[1]},
		AbuseContactHandles:   []rri.DenicHandle{handles[2]},
		NameServers:           nameServers,
	}, nil
}

func readContactData(args []string, dataOffset int) (rri.DenicHandle, rri.ContactData, error) {
	handle, err := rri.ParseDenicHandle(args[0])
	if err != nil {
		return rri.EmptyDenicHandle(), rri.ContactData{}, fmt.Errorf("%q: %s", args[0], err.Error())
	}

	console.Print("Type [PERSON ; ORG]> ")
	strContactType, err := commandline.ReadLineWithHistory(contactTypeHist())
	if err != nil {
		return rri.EmptyDenicHandle(), rri.ContactData{}, err
	}
	contactType, err := rri.ParseContactType(strContactType)
	if err != nil {
		return rri.EmptyDenicHandle(), rri.ContactData{}, fmt.Errorf("%q: %s", strContactType, err.Error())
	}

	inputDataLabels := []string{"Name", "Address", "Postal Code", "City", "Country Code", "E-Mail"}
	inputData := make([]string, 0)
	for i := 0; ; i++ {
		var label string
		if i >= len(inputDataLabels) {
			label = inputDataLabels[len(inputDataLabels)-1]
		} else {
			label = inputDataLabels[i]
		}
		console.Printf("%s> ", label)

		str, err := console.ReadLine()
		if err != nil {
			return rri.EmptyDenicHandle(), rri.ContactData{}, err
		}
		if len(str) == 0 && i >= 5 {
			break
		}
		inputData = append(inputData, str)
	}

	return handle, rri.ContactData{
		Type:        contactType,
		Name:        inputData[0],
		Address:     inputData[1],
		PostalCode:  inputData[2],
		City:        inputData[3],
		CountryCode: inputData[4],
		EMail:       inputData[5:],
	}, nil
}

func contactTypeHist() commandline.LineHistory {
	hist := commandline.NewLineHistory(2)
	hist.Put("ORG")
	hist.Put("PERSON")
	return hist
}
