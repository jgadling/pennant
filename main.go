package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/franela/goreq"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "pennant"
	app.Version = "1.2"
	app.Usage = "Tool to configure and check feature flags"
	app.Commands = []cli.Command{
		{
			Name:   "test",
			Usage:  "Test whether a flag is en/disabled based on a policy+document",
			Action: checkFlag,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "Show more output",
				},
				cli.StringFlag{
					Name:  "f, file",
					Usage: "Specify a flag config file",
				},
				cli.StringFlag{
					Name:  "d, datafile",
					Usage: "Specify a data file",
				},
			},
		},
		{
			Name:     "list",
			Category: "flag commands",
			Usage:    "List flags",
			Action:   listFlags,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "Return json output",
				},
			},
		},
		{
			Name:     "update",
			Category: "flag commands",
			Usage:    "Create or update a flag",
			Action:   updateFlag,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "f, file",
					Usage: "Specify a flag config file",
				},
				cli.BoolFlag{
					Name:  "json",
					Usage: "Return json output",
				},
			},
		},
		{
			Name:     "show",
			Category: "flag commands",
			Usage:    "Show details for a flag",
			Action:   showFlag,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "Return json output",
				},
			},
		},
		{
			Name:     "value",
			Category: "flag commands",
			Usage:    "Return whether a flag is en/disabled for a document",
			Action:   flagValue,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "Return json output",
				},
				cli.StringFlag{
					Name:  "d, document",
					Usage: "specify a document file",
				},
			},
		},
		{
			Name:     "delete",
			Category: "flag commands",
			Usage:    "Delete a flag",
			Action:   deleteFlag,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "Return json output",
				},
			},
		},
		{
			Name:   "server",
			Usage:  "Run pennant server",
			Action: runServer,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "Show more output",
				},
				cli.StringFlag{
					Name:  "c, conf",
					Value: "pennant.json",
					Usage: "Specify a config file",
				},
			},
		},
	}

	args := os.Args
	if err := app.Run(args); err != nil {
		logger.Criticalf("Error: %v", err)
		os.Exit(1)
	}
}

func runServer(c *cli.Context) error {
	logger.Warning("We should start a server here, eh?")
	conf, err := loadConfig(c.String("conf"))
	if err != nil {
		return fmt.Errorf("Couldn't load config file: %v", err)
	}
	logger.Infof("config %v", conf)
	driver, err := conf.getDriver()
	if err != nil {
		return fmt.Errorf("Problem loading driver %v", err)
	}
	logger.Infof("driver %v", driver)
	fc, _ := NewFlagCache()
	modifyIndex, err := driver.loadAllFlags(fc)
	if err != nil {
		return fmt.Errorf("Problem loading policies %v", err)
	}
	go driver.watchForChanges(fc, modifyIndex)

	go runGrpc(conf, fc)
	runHttp(conf, fc, driver)

	return nil
}

func listFlags(c *cli.Context) error {
	url := genUrl("/flags")
	var flagList FlagListResponse
	req := goreq.Request{Uri: url}
	err, complete := doRestRequest(c, req, &flagList)
	if err != nil || complete {
		return err
	}
	if len(flagList.Flags) == 0 {
		logger.Info("Sorry, no flags yet")
		return nil
	}
	cp := NewColPrinter([]string{"Name"}, "  ")
	for _, v := range flagList.Flags {
		cp.AddRow([]string{v})
	}
	cp.Print()
	return nil
}

func showFlag(c *cli.Context) error {
	flagName := c.Args().Get(0)
	if len(flagName) == 0 {
		return fmt.Errorf("a flag name argument is required")
	}
	url := genUrl(fmt.Sprintf("/flags/%s", flagName))
	req := goreq.Request{Uri: url}
	var flagResp FlagItemResponse
	err, complete := doRestRequest(c, req, &flagResp)
	if err != nil || complete {
		return err
	}
	prettyPrintFlag(flagResp.Flag)
	return nil
}

func updateFlag(c *cli.Context) error {
	flagData, err := loadFileFromCli(c, "file", 0)
	if err != nil {
		return err
	}
	url := genUrl("/flags")
	req := goreq.Request{Uri: url, Method: "POST", Body: flagData}
	var flagResp FlagItemResponse
	err, complete := doRestRequest(c, req, &flagResp)
	if err != nil || complete {
		return err
	}
	prettyPrintFlag(flagResp.Flag)
	return nil
}

func deleteFlag(c *cli.Context) error {
	flagName := c.Args().Get(0)
	json := c.Bool("json")
	if len(flagName) == 0 {
		return fmt.Errorf("a flag name argument is required")
	}
	url := genUrl(fmt.Sprintf("/flags/%s", flagName))
	res, err := goreq.Request{Uri: url, Method: "DELETE"}.Do()
	if err != nil {
		return err
	}
	if json {
		body, _ := res.Body.ToString()
		fmt.Print(body)
		return nil
	}
	var flagResp FlagValueResponse
	res.Body.FromJsonTo(&flagResp)
	if flagResp.Status != 200 {
		return fmt.Errorf("%d - %s", flagResp.Status, flagResp.Message)
	}
	fmt.Printf("%s deleted\n", flagName)
	return nil
}

func flagValue(c *cli.Context) error {
	flagName := c.Args().Get(0)
	document, err := loadFileFromCli(c, "document", 1)
	if err != nil {
		return err
	}
	json := c.Bool("json")
	if len(flagName) == 0 {
		return fmt.Errorf("flag name is required")
	}
	url := genUrl(fmt.Sprintf("/flagValue/%s", flagName))
	res, err := goreq.Request{Uri: url, Method: "POST", Body: document}.Do()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	if json {
		body, _ := res.Body.ToString()
		fmt.Print(body)
		return nil
	}
	var valueResp FlagValueResponse
	res.Body.FromJsonTo(&valueResp)
	if valueResp.Status != 200 {
		return fmt.Errorf("%d - %s", valueResp.Status, valueResp.Message)
	}
	enabled := "disabled"
	if valueResp.Enabled == true {
		enabled = "enabled"
	}
	cp := NewColPrinter([]string{"Flag", "Status"}, "  ")
	cp.AddRow([]string{
		flagName,
		enabled,
	})
	cp.Print()
	return nil
}

func checkFlag(c *cli.Context) error {
	flagfile, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("can't read file %s", c.String("file"))
	}
	flag, err := LoadAndParseFlag(flagfile)
	if err != nil {
		return err
	}
	datafile, err := ioutil.ReadFile(c.String("datafile"))
	if err != nil {
		return fmt.Errorf("Can't read file %s", c.String("datafile"))
	}
	datas := make(map[string]interface{})
	if err = json.Unmarshal(datafile, &datas); err != nil {
		return fmt.Errorf("Can't parse file %s", c.String("datafile"))
	}

	flag.Parse()
	logger.Infof("Policy data is %v", flag.GetValue(datas))
	return nil
}


// Helpers below

func genUrl(url string) string {
	server := os.Getenv("PENNANT_SERVER")
	if len(server) == 0 {
		server = "http://127.0.0.1:1234"
	}
	return fmt.Sprintf("%s%s", server, url)
}

func prettyPrintFlag(flag *Flag) {
	cp := NewColPrinter([]string{"Name", "Description", "DefaultValue"}, "  ")
	cp.AddRow([]string{
		flag.Name,
		flag.Description,
		strconv.FormatBool(flag.DefaultValue)})
	cp.Print()
	fmt.Println()

	cp = NewColPrinter([]string{"Rule", "Comment"}, "  ")
	for _, v := range flag.Policies {
		cp.AddRow([]string{
			v.Rules,
			v.Comment})
	}
	cp.Print()
}

func doRestRequest(c *cli.Context, req goreq.Request, container interface{}) (error, bool) {
	res, err := req.Do()
	if err != nil {
		return err, false
	}
	printJson := c.Bool("json")
	if printJson {
		body, _ := res.Body.ToString()
		fmt.Print(body)
		return nil, true
	}
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		var respData map[string]interface{}
        json.Unmarshal(body, &respData)
		return fmt.Errorf("%d - %v", res.StatusCode, respData["message"]), false
	}
	res.Body.FromJsonTo(container)
	return nil, false
}

func loadFileFromCli(c *cli.Context, fieldName string, argIndex int) (string, error) {
	document := c.Args().Get(argIndex)
	if document == "-" {
		if stdinData, err := ioutil.ReadAll(os.Stdin); err != nil {
		    return document, err
	    } else {
		    return string(stdinData), nil
	    }
	}
	if len(document) > 0 {
		return document, nil
	}
	docFile := c.String(fieldName)
	if len(docFile) == 0 {
		return document, fmt.Errorf("a flag file name or literal flag argument is required")
	}
	fileData, err := ioutil.ReadFile(docFile)
	if err != nil || len(fileData) < 2 {
		return document, fmt.Errorf("can't read file %s", docFile)
	}
	return  string(fileData), nil
}
