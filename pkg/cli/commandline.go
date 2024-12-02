package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DENICeG/go-rriclient/internal/env"
	"github.com/DENICeG/go-rriclient/pkg/preset"
	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/commandline"
)

type Service struct {
	rriClient                  *rri.Client
	completion                 *domainOrHandleCompletion
	ReturnErrorOnFail          bool
	customCommands             []customCommand
	colorPromptRRI             string
	colorPromptUser            string
	colorPromptHost            string
	colorSendRaw               string
	colorReceiveRaw            string
	colorSuccessResponse       string
	colorErrorResponseMessage  string
	colorTechnicalErrorMessage string
	colorInnerError            string
	colorEnd                   string
	signSend                   string
	signReceive                string
	presets                    preset.Data
	embedFS                    embed.FS
}

// New returns a new Service instance.
func New(client *rri.Client, presets preset.Data, embedFS embed.FS) *Service {
	result := &Service{
		rriClient:                  client,
		completion:                 NewCompletion(),
		colorPromptRRI:             "\033[1;34m",
		colorPromptUser:            "\033[1;32m",
		colorPromptHost:            "\033[1;32m",
		colorSendRaw:               "\033[0;94m",
		colorReceiveRaw:            "\033[0;96m",
		colorSuccessResponse:       "\033[0;29m",
		colorErrorResponseMessage:  "\033[0;91m",
		colorTechnicalErrorMessage: "\033[1;91m",
		colorInnerError:            "\033[2;91m",
		colorEnd:                   "\033[0m",
		signSend:                   "-->",
		signReceive:                "<--",
		presets:                    presets,
		embedFS:                    embedFS,
	}

	if !console.SupportsColors() {
		result.disableColors()
	}

	return result
}

func (s *Service) disableColors() {
	s.colorPromptRRI = ""
	s.colorPromptUser = ""
	s.colorPromptHost = ""
	s.colorSendRaw = ""
	s.colorReceiveRaw = ""
	s.colorSuccessResponse = ""
	s.colorErrorResponseMessage = ""
	s.colorTechnicalErrorMessage = ""
	s.colorInnerError = ""
	s.colorEnd = ""
}

func (s *Service) PrintColorsAndSigns() {
	printColor := func(name, colorStr string) {
		console.Printlnf("  %s%s%s", colorStr, name, s.colorEnd) //nolint
	}
	printSign := func(name, sign string) {
		console.Printlnf("  %s: %q", name, sign) //nolint
	}

	printColor("ColorDefault", s.colorEnd)
	printColor("ColorPromptRRI", s.colorPromptRRI)
	printColor("ColorPromptUser", s.colorPromptUser)
	printColor("ColorPromptHost", s.colorPromptHost)
	printColor("ColorSendRaw", s.colorSendRaw)
	printColor("ColorReceiveRaw", s.colorReceiveRaw)
	printColor("ColorSuccessResponse", s.colorSuccessResponse)
	printColor("ColorErrorResponseMessage", s.colorErrorResponseMessage)
	printColor("ColorTechnicalErrorMessage", s.colorTechnicalErrorMessage)
	printColor("ColorInnerError", s.colorInnerError)
	printSign("SignSend", s.signSend)
	printSign("SignReceive", s.signReceive)
}

func (s *Service) Run(confDir string, cmd []string) error {
	var err error
	s.customCommands, err = readCustomCommands(filepath.Join(confDir, "custom-commands"))
	if err != nil {
		if !os.IsNotExist(err) {
			console.Println("Failed to import custom commands:", err.Error())
		}
	}

	cli := s.prepareCLI()

	if len(cmd) > 0 {
		// exec command that has been passed via command line and return result
		return cli.ExecCommand(cmd[0], cmd[1:])
	}

	console.Println("Interactive RRI Command Line")
	console.Println("  type 'help' to see a list of available commands")
	console.Println("  use tab for auto-completion and arrow keys for history")

	// start interactive command line loop
	if err := cli.Run(); err != nil {
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
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := make([]customCommand, 0)
	for _, f := range files {
		if !f.IsDir() {
			data, err := os.ReadFile(filepath.Join(dir, f.Name()))
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
			result = append(result, cmd)
		}
	}

	return result, nil
}

func (s *Service) prepareCLI() *commandline.Environment {
	cli := commandline.NewEnvironment()
	cli.Prompt = func() string {
		var prefix, user, host, suffix string
		prefix = fmt.Sprintf("%sRRI%s", s.colorPromptRRI, s.colorEnd)
		if s.rriClient.IsLoggedIn() {
			user = fmt.Sprintf("%s%s%s@", s.colorPromptUser, s.rriClient.CurrentUser(), s.colorEnd)
		}
		host = fmt.Sprintf("%s%s%s", s.colorPromptHost, s.rriClient.RemoteAddress(), s.colorEnd)
		return fmt.Sprintf("%s{%s%s}%s", prefix, user, host, suffix)
	}

	cli.ErrorHandler = func(_ string, _ []string, err error) error {
		console.Printlnf("%sERROR: %s%s", s.colorTechnicalErrorMessage, err.Error(), s.colorEnd) //nolint
		return nil
	}

	cli.RegisterCommand(commandline.NewExitCommand("exit"))
	cli.RegisterCommand(commandline.NewCustomCommand("help", nil, s.cmdHelp))

	cli.RegisterCommand(commandline.NewCustomCommand("login", nil, s.cmdLogin))
	cli.RegisterCommand(commandline.NewCustomCommand("logout", nil, s.cmdLogout))

	s.registerSwitchCommand(cli, "create", cmdSwitches{
		Domain:    s.cmdCreateDomain,
		Handle:    s.cmdCreateHandle,
		AuthInfo1: s.cmdCreateAuthInfo1,
	})
	s.registerSwitchCommand(cli, "check", cmdSwitches{
		Domain: s.newDomainQueryCommand(rri.NewCheckDomainQuery),
		Handle: s.newHandleQueryCommand(rri.NewCheckHandleQuery),
	})
	s.registerSwitchCommand(cli, "info", cmdSwitches{
		Domain: s.newDomainQueryCommand(rri.NewInfoDomainQuery),
		Handle: s.newHandleQueryCommand(rri.NewInfoHandleQuery),
	})
	s.registerSwitchCommand(cli, "update", cmdSwitches{
		Domain: s.cmdUpdateDomain,
	})

	s.registerDomainCommand(cli, "delete", s.newDomainQueryCommand(rri.NewDeleteDomainQuery))
	s.registerDomainCommand(cli, "restore", s.newDomainQueryCommand(rri.NewRestoreDomainQuery))
	s.registerDomainCommand(cli, "transit", s.cmdTransit, commandline.NewOneOfArgCompletion("disconnect", "connect"))
	s.registerDomainCommand(cli, "chholder", s.cmdChangeHolder)
	s.registerDomainCommand(cli, "chprov", s.cmdChangeProvider)

	s.registerDomainCommand(cli, "queue-read", s.cmdQueueRead)
	s.registerDomainCommand(cli, "queue-delete", s.cmdQueueDelete)

	// register custom commands
	for _, cmd := range s.customCommands {
		s.registerCustomCommand(cli, cmd)
	}

	cli.RegisterCommand(commandline.NewCustomCommand("raw", nil, s.cmdRaw))
	cli.RegisterCommand(commandline.NewCustomCommand("file", commandline.NewFixedArgCompletion(commandline.NewLocalFileSystemArgCompletion(true)), s.HandleFile))

	cli.RegisterCommand(commandline.NewCustomCommand("verbose", nil, s.cmdVerbose))
	cli.RegisterCommand(commandline.NewCustomCommand("preset", nil, s.HandlePreset))

	return cli
}

func (s *Service) cmdHelp(args []string) error {
	type customCmd struct {
		Desc string
		Cmd  []string
		Args []string
	}
	commands := []customCmd{
		{Cmd: []string{"exit"}, Args: nil, Desc: "exit application"},
		{Cmd: []string{"help"}, Args: nil, Desc: "show this help"},
		{},
		{Cmd: []string{"login"}, Args: []string{"user", "password"}, Desc: "log in to a RRI account"},
		{Cmd: []string{"logout"}, Args: nil, Desc: "log out from the current RRI account"},
		{},
		{Cmd: []string{"create", "handle"}, Args: []string{"domain"}, Desc: "send a CREATE command for a specific handle"},
		{Cmd: []string{"check", "handle"}, Args: []string{"domain"}, Desc: "send a CHECK command for a specific handle"},
		{Cmd: []string{"info", "handle"}, Args: []string{"domain"}, Desc: "send an INFO command for a specific handle"},
		// TODO contact-update
		{},
		{Cmd: []string{"create", "domain"}, Args: []string{"domain"}, Desc: "send a CREATE command for a new domain"},
		{Cmd: []string{"check", "domain"}, Args: []string{"domain"}, Desc: "send a CHECK command for a specific domain"},
		{Cmd: []string{"info", "domain"}, Args: []string{"domain"}, Desc: "send an INFO command for a specific domain"},
		{Cmd: []string{"update", "domain"}, Args: []string{"domain"}, Desc: "send an UPDATE command for a specific domain"},
		{Cmd: []string{"chholder", "domain"}, Args: []string{"domain"}, Desc: "send an CHHOLDER command for a specific domain"},
		{},
		{Cmd: []string{"delete"}, Args: []string{"domain"}, Desc: "send a DELETE command for a specific domain"},
		{Cmd: []string{"restore"}, Args: []string{"domain"}, Desc: "send a RESTORE command for a specific domain"},
		{Cmd: []string{"transit"}, Args: []string{"domain"}, Desc: "send a TRANSIT command for a specific domain"},
		{Cmd: []string{"create", "authinfo1"}, Args: []string{"domain", "secret", "expire"}, Desc: "send a CREATE-AUTHINFO1 command for a specific domain"},
		// TODO create-authinfo2
		// TODO delete-authinfo1
		{Cmd: []string{"chprov"}, Args: []string{"domain", "secret"}, Desc: "send a CHPROV command for a specific domain"},
		{},
		{Cmd: []string{"verify-queue-read"}, Args: nil, Desc: "send a VERIFY-QUEUE-READ command"},
		{Cmd: []string{"verify-queue-delete"}, Args: []string{"msgid"}, Desc: "send a VERIFY-QUEUE-DELETE command for a specific vChecked message"},
		// TODO verify
		// -
		// TODO queue-read
		// TODO queue-delete
		// TODO regacc-info
		{},
		{Cmd: []string{"raw"}, Args: nil, Desc: "enter a raw query and send it"},
		{Cmd: []string{"file"}, Args: []string{"path"}, Desc: "process a query file as accepted by flag --file"},
		{},
		{Cmd: []string{"verbose"}, Args: nil, Desc: "toggle verbose mode"},
		{},
		{Cmd: []string{"preset"}, Args: []string{"pat"}, Desc: "Execute a preset, that can be edited by the user"},
	}

	if len(s.customCommands) > 0 {
		head := commands[:25]
		tail := make([]customCmd, 6)
		copy(tail, commands[25:])
		commands = head
		for _, cmd := range s.customCommands {
			args := make([]string, 0)
			for _, arg := range cmd.Args {
				if arg.IsInputParameter() {
					args = append(args, arg.Name)
				}
			}
			commands = append(commands, customCmd{Cmd: []string{cmd.Cmd}, Args: args, Desc: cmd.Description})
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

type cmdSwitches struct {
	Domain    commandline.ExecCommandHandler
	Handle    commandline.ExecCommandHandler
	AuthInfo1 commandline.ExecCommandHandler
}

func (s *Service) registerSwitchCommand(cle *commandline.Environment, name string, switches cmdSwitches) {
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
			s.completion,
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

func (s *Service) registerDomainCommand(cle *commandline.Environment, name string, commandHandler commandline.ExecCommandHandler, additionalCompletionHandlers ...commandline.ArgCompletion) {
	argCompletionHandlers := []commandline.ArgCompletion{s.completion.histDomains}
	argCompletionHandlers = append(argCompletionHandlers, additionalCompletionHandlers...)
	cle.RegisterCommand(commandline.NewCustomCommand(name, commandline.NewFixedArgCompletion(argCompletionHandlers...), commandHandler))
}

func (s *Service) newDomainQueryCommand(f func(domain string) *rri.Query) commandline.ExecCommandHandler {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing domain name")
		}

		_, err := s.processQuery(f(args[0]))
		s.completion.PutDomain(args[0])
		return err
	}
}

func (s *Service) newHandleQueryCommand(f func(handle rri.DenicHandle) *rri.Query) commandline.ExecCommandHandler {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing handle")
		}

		handle, err := rri.ParseDenicHandle(args[0])
		if err != nil {
			return err
		}

		_, err = s.processQuery(f(handle))
		s.completion.PutHandle(args[0])
		return err
	}
}

func (s *Service) registerCustomCommand(cli *commandline.Environment, cmd customCommand) {
	clArgs := make([]commandline.ArgCompletion, 0)
	for _, arg := range cmd.Args {
		if strings.EqualFold(arg.Type, "domain") {
			clArgs = append(clArgs, s.completion.histDomains)
		}
	}

	cli.RegisterCommand(commandline.NewCustomCommand(cmd.Cmd, commandline.NewFixedArgCompletion(clArgs...), func(args []string) error {
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
		// TODO fill history
		query := rri.NewQuery(rri.LatestVersion, rri.QueryAction(cmd.Action), fields, nil)
		_, err := s.processQuery(query)
		return err
	}))
}

func (s *Service) RawQueryPrinter(msg string, isOutgoing bool) {
	if isOutgoing {
		console.Printlnf("%s %s%q%s", s.signSend, s.colorSendRaw, rri.CensorRawMessage(msg), s.colorEnd)
	} else {
		console.Printlnf("%s %s%q%s", s.signReceive, s.colorReceiveRaw, rri.CensorRawMessage(msg), s.colorEnd)
	}
}

func (s *Service) ErrorPrinter(err error) {
	console.Printlnf("%sERR: %s%s", s.colorInnerError, err.Error(), s.colorEnd)
}

func (s *Service) processQuery(query *rri.Query) (bool, error) {
	response, err := s.rriClient.SendQuery(query)
	if err != nil {
		return false, fmt.Errorf("failed to send query: %w", err)
	}

	var colorStr string
	if response.IsSuccessful() {
		colorStr = s.colorSuccessResponse
	} else {
		colorStr = s.colorErrorResponseMessage
	}

	console.Print(colorStr)
	console.Println(response.String())
	console.Print(s.colorEnd)

	if s.ReturnErrorOnFail && !response.IsSuccessful() {
		return false, fmt.Errorf("RRI returned result 'failed'")
	}

	return response.IsSuccessful(), nil
}

func (s *Service) readDomainData(args []string, dataOffset int) (string, rri.DomainData, error) {
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
			str, err := commandline.ReadLineWithHistory(s.completion.histHandles)
			if err != nil {
				return "", rri.DomainData{}, err
			}

			handle, err := rri.ParseDenicHandle(str)
			if err != nil {
				return "", rri.DomainData{}, fmt.Errorf("%q: %s", str, err.Error())
			}

			s.completion.PutHandle(str)
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

func readContactData(args []string) (rri.DenicHandle, rri.ContactData, error) {
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

func RetrieveEnvironment(envReader *env.Reader, host, environment, user, password *string, cmd *[]string) (env.Environment, error) {
	var addressFromCommandLine string
	if len(*host) > 0 {
		addressFromCommandLine = *host
		if !strings.Contains(addressFromCommandLine, ":") {
			// did the user forget to specify a port? use default port
			addressFromCommandLine += ":51131"
		}

	} else if len(*cmd) >= 1 && strings.Contains((*cmd)[0], ":") {
		// consume first command part as address for backwards compatibility
		addressFromCommandLine = (*cmd)[0]
		*cmd = (*cmd)[1:]
	}

	var envi env.Environment
	if len(*environment) > 0 {
		err := envReader.CreateOrReadEnvironment(*environment, &envi)
		if err != nil {
			return env.Environment{}, err
		}
	} else if len(addressFromCommandLine) == 0 {
		err := envReader.SelectEnvironment(&envi)
		if err != nil {
			return env.Environment{}, err
		}
	}

	if len(addressFromCommandLine) > 0 {
		envi.Address = addressFromCommandLine
	}
	if len(*user) > 0 {
		envi.User = *user
	}
	if len(*password) > 0 {
		envi.Password = *password
	}

	if len(envi.User) > 0 && len(envi.Password) == 0 {
		// ask for missing user credentials
		var err error
		console.Printlnf("Please enter RRI password for user %q", envi.User)
		console.Print("> ")
		envi.Password, err = console.ReadPassword()
		if err != nil {
			return env.Environment{}, err
		}
	}

	return envi, nil
}

func EnterEnvironment(envi any) error {
	e, ok := envi.(*env.Environment)
	if !ok {
		panic(fmt.Sprintf("environment has unexpected type %T", envi))
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

func DeleteEnv(argDeleteEnv *string, envReader *env.Reader) (bool, error) {
	if len(*argDeleteEnv) == 0 {
		return false, nil
	}

	err := envReader.DeleteEnvironment(*argDeleteEnv)
	if err != nil {
		return false, err
	}

	console.Printlnf("environment %q has been deleted", *argDeleteEnv)
	return true, nil // should exit after deleting env
}

func PrintEnv(argListEnv *bool, envReader *env.Reader) (bool, error) {
	if !*argListEnv {
		return false, nil
	}

	environments, err := envReader.ListEnvironments() //nolint
	if err != nil {
		return false, err
	}

	for _, env := range environments {
		console.Printlnf("- %s", env)
	}

	return true, nil // should exit after printing env
}

func PrintVersion(argVersion *bool, buildTime, gitCommit, version string) bool {
	if !*argVersion {
		return false
	}

	console.Printlnf("Standalone RRI Client v%s", version)
	if len(buildTime) > 0 && len(gitCommit) > 0 {
		console.Printlnf("  built at %s from commit %s", buildTime, gitCommit)
	}

	return true // should exit after printing version
}
