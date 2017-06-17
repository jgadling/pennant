package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
)

func main() {
	pennantCli()
}

func pennantCli() {
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
		os.Exit(1)
	}
}

func runServer(c *cli.Context) error {
	logger.Warning("We should start a server here, eh?")
	conf, err := loadConfig(c.String("conf"))
	if err != nil {
		logger.Criticalf("Couldn't load config file: %v", err)
	}
	logger.Infof("config %v", conf)
	driver, err := conf.getDriver()
	if err != nil {
		logger.Errorf("Problem loading driver %v", err)
	}
	logger.Infof("driver %v", driver)
	fc, _ := NewFlagCache()
	modifyIndex, err := driver.loadAllFlags(fc)
	if err != nil {
		logger.Warningf("Problem loading policies %v", err)
		return err
	}
	go driver.watchForChanges(fc, modifyIndex)

	go runGrpc(conf, fc)
	runHttp(conf, fc, driver)

	return nil
}

func checkFlag(c *cli.Context) error {
	logger.Info("running checkflag")
	flagfile, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		logger.Criticalf("can't read file %s", c.String("file"))
		return nil
	}
	flag, err := LoadAndParseFlag(flagfile)
	if err != nil {
		logger.Criticalf("can't parse file %s", c.String("file"))
		return err
	}
	logger.Infof("flag is %v", flag)
	datafile, err := ioutil.ReadFile(c.String("datafile"))
	if err != nil {
		logger.Criticalf("Can't read file %s", c.String("datafile"))
		return nil
	}
	datas := make(map[string]interface{})
	if err = json.Unmarshal(datafile, &datas); err != nil {
		logger.Criticalf("Can't parse file %s", c.String("datafile"))
		return err
	}

	logger.Info("Parsing!")
	flag.Parse()
	logger.Infof("Policy data is %v", flag.GetValue(datas))
	return nil
}
