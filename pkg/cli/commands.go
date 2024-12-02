package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/parser"
	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-console/input"
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

		s.printXMLResult(resp, i)
	}

	return nil
}

func (s *Service) printXMLResult(resp string, i int) {
	isSuccess := strings.Contains(resp, "<tr:result>success</tr:result>")

	console.Println("----------------------------------------")
	console.Println(fmt.Sprintf("Query #%v has success result: %v", i+1, isSuccess))
	console.Println("----------------------------------------")
	console.Println(resp)
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

// lines := parser.SplitLines(data)
// queryStrings := parser.SplitQueries(lines)

// if isXMLFile {
// 	s.
// 	return nil
// }

// err = s.executeKVQueries(queryStrings)
// if err != nil {
// 	return err
// }

// return nil

func (s *Service) cmdXML(args []string) error {
	s.rriClient.XMLMode = !s.rriClient.XMLMode
	if s.rriClient.XMLMode {
		console.Println("XML mode on")
	} else {
		console.Println("XML mode off")
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

func (s *Service) CmdDisplayPresets(args []string) error {
	console.Println("Choose on of the possible presets.")
	counter := 0
	for key, orderType := range s.presets {
		console.Printlnf("%v", key)
		for _, preset := range orderType {
			console.Printlnf("\t%v %v", counter, preset)
			counter++
		}
	}
	return nil
}
