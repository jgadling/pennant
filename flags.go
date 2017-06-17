package main

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
)

type Flag struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	DefaultValue bool     `json:"default"`
	Policies     []Policy `json:"policies"`
	Version      uint64   `json:"-"` // Yeah this abstraction is leaky :(
}

type Policy struct {
	Comment    string                         `json:"comment"`
	Rules      string                         `json:"rules"`
	ParsedExpr *govaluate.EvaluableExpression `json:"-"`
}

func LoadFlagJson(flagData []byte) (*Flag, error) {
	flag := Flag{}
	if err := json.Unmarshal(flagData, &flag); err != nil {
		return &flag, fmt.Errorf("can't parse flag %s", flagData)
	}
	return &flag, nil
}

func LoadAndParseFlag(flagData []byte) (*Flag, error) {
	flag := Flag{}
	if err := json.Unmarshal(flagData, &flag); err != nil {
		return &flag, fmt.Errorf("can't parse flag %s", flagData)
	}
	err := flag.Parse()
	return &flag, err
}

func (f *Flag) Parse() error {
	logger.Infof("loading %v", f)
	boringExpr, _ := govaluate.NewEvaluableExpressionWithFunctions("false", getLibraryFunctions())
	for i := range f.Policies {
		policy := &f.Policies[i]

		logger.Infof("loading %v", policy.Rules)

		expr, err := govaluate.NewEvaluableExpressionWithFunctions(policy.Rules, getLibraryFunctions())
		if err != nil {
			policy.ParsedExpr = boringExpr
			return err
		}
		logger.Infof("parsed expr %v", expr)
		policy.ParsedExpr = expr
	}
	return nil
}

func (f *Flag) GetValue(params map[string]interface{}) bool {
	logger.Infof("getting value %v", params)
	returnval := f.DefaultValue
	if returnval {
		return returnval
	}
	messages := make(chan bool)
	for i := range f.Policies {
		go func(policy *Policy) {
			if policy.ParsedExpr == nil {
				logger.Errorf("Something is nullified")
				messages <- false
				return
			}
			res, err := policy.ParsedExpr.Evaluate(params)
			if err != nil {
				logger.Errorf("err is %v", err)
				messages <- false
				return
			}
			logger.Infof("result is %v", res)
			if res == true {
				messages <- true
			}
			messages <- false
		}(&f.Policies[i])
	}
	// Wait for responses.
	for i := 0; i < len(f.Policies); i++ {
		if <-messages {
			return true
		}
	}
	return false
}
