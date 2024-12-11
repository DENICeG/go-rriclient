package cli

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/highlight"

	"github.com/DENICeG/go-console/v2"
	"github.com/DENICeG/go-console/v2/input"
	"github.com/DENICeG/go-rriclient/pkg/parser"
	"github.com/DENICeG/go-rriclient/pkg/preset"
	"github.com/DENICeG/go-rriclient/pkg/rri"
)

func (s *Service) cmdLogin(args []string) error {
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

	if s.rriClient.IsLoggedIn() {
		console.Println("Exiting current session")
		s.rriClient.Logout()
	}

	return s.rriClient.Login(args[0], pass)
}

func (s *Service) cmdLogout(args []string) error {
	return s.rriClient.Logout()
}

func (s *Service) cmdCreateHandle(args []string) error {
	handle, contactData, err := readContactData(args)
	if err != nil {
		return err
	}

	_, err = s.processQuery(rri.NewCreateContactQuery(handle, contactData))
	s.completion.PutHandle(handle.String())

	return err
}

func (s *Service) cmdCreateDomain(args []string) error {
	domainName, domainData, err := s.readDomainData(args, 1)
	if err != nil {
		return err
	}

	_, err = s.processQuery(rri.NewCreateDomainQuery(domainName, domainData))
	s.completion.PutDomain(domainName)

	return err
}

func (s *Service) cmdUpdateDomain(args []string) error {
	domainName, domainData, err := s.readDomainData(args, 1)
	if err != nil {
		return err
	}

	// TODO use old domain values for empty fields -> only change explicitly entered data

	_, err = s.processQuery(rri.NewUpdateDomainQuery(domainName, domainData))
	s.completion.PutDomain(domainName)

	return err
}

func (s *Service) cmdChangeHolder(args []string) error {
	domainName, domainData, err := s.readDomainData(args, 1)
	if err != nil {
		return err
	}

	// TODO use old domain values for empty fields -> only change explicitly entered data

	_, err = s.processQuery(rri.NewChangeHolderQuery(domainName, domainData))
	s.completion.PutDomain(domainName)

	return err
}

func (s *Service) cmdTransit(args []string) error {
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

	_, err := s.processQuery(rri.NewTransitDomainQuery(domainName, disconnect))
	s.completion.PutDomain(domainName)

	return err
}

func (s *Service) cmdCreateAuthInfo1(args []string) error {
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

	_, err := s.processQuery(rri.NewCreateAuthInfo1Query(args[0], args[1], expire))
	return err
}

func (s *Service) cmdChangeProvider(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing domain name")
	}

	if len(args) < 2 {
		return fmt.Errorf("missing auth info secret")
	}

	domainName, domainData, err := s.readDomainData(args, 2)
	if err != nil {
		return err
	}

	_, err = s.processQuery(rri.NewChangeProviderQuery(domainName, args[1], domainData))
	s.completion.PutDomain(domainName)

	return err
}

func (s *Service) cmdQueueRead(args []string) error {
	_, err := s.processQuery(rri.NewQueueReadQuery(""))
	return err
}

func (s *Service) cmdQueueDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing message id")
	}

	_, err := s.processQuery(rri.NewQueueDeleteQuery(args[0], ""))
	return err
}

func (s *Service) cmdRaw(args []string) error {
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
		response, err := s.rriClient.SendRaw(rawCommand)
		if err != nil {
			return err
		}

		if strings.Contains(response, "<") {
			response, err = highlight.Transform(response, highlight.XML)
			if err != nil {
				return err
			}
		} else {
			response, err = highlight.Transform(response, highlight.YAML)
			if err != nil {
				return err
			}
		}

		console.Println(response)

		if s.ReturnErrorOnFail {
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

func (s *Service) HandleFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing query file")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	lines := parser.SplitLines(data)
	queries := parser.SplitQueries(lines)

	if isXML(data) {
		return s.executeXMLQueries(queries)
	}

	err = s.executeKVQueries(queries)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) executeXMLQueries(queries []string) error {
	for i, query := range queries {
		resp, err := s.rriClient.SendRaw(query)
		if err != nil {
			return err
		}

		err = s.printXMLResult(resp, i, highlight.XML)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) printXMLResult(resp string, i int, format highlight.Format) error {
	isSuccess := strings.Contains(resp, "<tr:result>success</tr:result>")

	console.Println("----------------------------------------")
	console.Println(fmt.Sprintf("Query #%v has success result: %v", i+1, isSuccess))
	console.Println("----------------------------------------")

	resp, err := highlight.Transform(resp, format)
	if err != nil {
		return err
	}

	console.Println(resp)
	// println(resp)

	return nil
}

func isXML(data []byte) bool {
	return bytes.Contains(data, []byte("<"))
}

func (s *Service) executeKVQueries(queryStrings []string) error {
	skipAuthQueries := s.rriClient.IsLoggedIn()

	queries, err := parser.ParseQueriesKV(queryStrings)
	if err != nil {
		return err
	}

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
		// TODO colored orange
		console.Println("Currently logged in. Auth queries will be skipped")
	}

	for _, query := range queries {
		if skipAuthQueries {
			if query.Action() == rri.ActionLogin || query.Action() == rri.ActionLogout {
				continue
			}
		}

		success, err := s.processQuery(query)
		if err != nil {
			return err
		}

		if !success {
			break
		}
	}

	return nil
}

func (s *Service) cmdVerbose(args []string) error {
	if s.rriClient.RawQueryPrinter == nil {
		s.rriClient.RawQueryPrinter = s.RawQueryPrinter
		s.rriClient.InnerErrorPrinter = s.ErrorPrinter
		console.Println("Verbose mode on")
	} else {
		s.rriClient.RawQueryPrinter = nil
		s.rriClient.InnerErrorPrinter = nil
		console.Println("Verbose mode off")
	}

	return nil
}

func (s *Service) HandlePreset(args []string) error {
	var chosenPreset *preset.Entry
	var err error

	if len(args) > 0 {
		chosenPreset = s.presets.Get(args[0])
		if chosenPreset == nil {
			console.Println("Chosen preset not found")
			return nil
		}
	}

	if chosenPreset == nil {
		chosenPreset, err = s.manualPresetFlow()
		if err != nil {
			return err
		}
	}

	presetContent, err := s.embedFS.ReadFile("examples/" + chosenPreset.Type + "/" + chosenPreset.DirName + "/" + chosenPreset.FileName)
	if err != nil {
		return err
	}

	format := highlight.YAML
	if strings.EqualFold(chosenPreset.Type, "xml") {
		format = highlight.XML
	}

	// Highlight in edit mode doesn't work yet
	// stringContent, err := highlight.Transform(string(presetContent), format)
	// if err != nil {
	// 	return err
	// }

	result, _, err := input.Text(string(presetContent))
	if err != nil {
		return err
	}

	res, err := s.rriClient.SendRaw(result)
	if err != nil {
		return err
	}

	res, err = highlight.Transform(res, format)
	if err != nil {
		return err
	}

	console.Println(res) //nolint

	return err
}

func (s *Service) manualPresetFlow() (*preset.Entry, error) {
	format, err := s.chosePresetType()
	if err != nil {
		return nil, err
	}

	ClearTerminal()

	s.listPresets(format)
	chosenIndex, err := s.chosePreset()
	if err != nil {
		return nil, err
	}

	ClearTerminal()

	result := s.presets.Preset[chosenIndex]

	return &result, nil
}

func (s *Service) chosePreset() (int, error) {
	console.Println()
	console.Print("Preset number: ")
	numberString, err := console.ReadLine()
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(numberString)
}

func (s *Service) chosePresetType() (string, error) {
	format := ""
	var err error

	for {
		console.Println("Choose preset: \n1: kv  \n2: xml")
		format, err = console.ReadLine()
		if err != nil {
			return format, err
		}

		if format == "1" {
			return "kv", nil

		}

		if format == "2" {
			return "xml", nil
		}
	}
}

func (s *Service) listPresets(format string) {
	if format == "kv" {
		console.Print("kv")
	}

	var currentDirName string

	for i := 0; i < len(s.presets.Preset); i++ {
		if format != s.presets.Preset[i].Type {
			continue
		}

		if i == s.presets.XMLStartIndex {
			console.Print("xml")
		}

		if currentDirName != s.presets.Preset[i].DirName {
			currentDirName = s.presets.Preset[i].DirName
			console.Printf("\n\t%v: ", currentDirName)
		}

		console.Printf("\n\t\t%v %v", i, s.presets.Preset[i].FileName)
	}
}
