package main

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
)

// A feature flag is some metadata and a collection of policies
type Flag struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	DefaultValue bool     `json:"default"`
	Policies     []Policy `json:"policies"`
	Version      uint64   `json:"-"` // Yeah this abstraction is leaky :(
}

// A policy is a govaluate-compatible expression that returns true or false
type Policy struct {
	Comment    string                         `json:"comment"`
	Rules      string                         `json:"rules"`
	ParsedExpr *govaluate.EvaluableExpression `json:"-"`
}

// Load a flag but don't parse the policies yet
func LoadFlagJson(flagData []byte) (*Flag, error) {
	flag := Flag{}
	if err := json.Unmarshal(flagData, &flag); err != nil {
		return &flag, fmt.Errorf("can't parse flag %s", flagData)
	}
	return &flag, nil
}

// Load a flag and parse the policy expressions
func LoadAndParseFlag(flagData []byte) (*Flag, error) {
	flag := Flag{}
	if err := json.Unmarshal(flagData, &flag); err != nil {
		return &flag, fmt.Errorf("can't parse flag %s", flagData)
	}
	err := flag.Parse()
	return &flag, err
}

// Parse and cache all the policy expressions for this flag. This needs to be
// done before GetResult can be invoked.
func (f *Flag) Parse() error {
	fallbackExpr, _ := govaluate.NewEvaluableExpressionWithFunctions("false", getLibraryFunctions())
	for i := range f.Policies {
		policy := &f.Policies[i]

		expr, err := govaluate.NewEvaluableExpressionWithFunctions(policy.Rules, getLibraryFunctions())
		if err != nil {
			policy.ParsedExpr = fallbackExpr
			return err
		}
		policy.ParsedExpr = expr
	}
	return nil
}

// GetValue compares a document to the flag policies and returns the boolean
// result of the evaluation.
func (f *Flag) GetValue(params map[string]interface{}) bool {
	returnval := f.DefaultValue
	if returnval {
		return returnval
	}
	messages := make(chan bool)
	for i := range f.Policies {
		go func(policy *Policy) {
			if policy.ParsedExpr == nil {
				logger.Errorf("null value exception")
				messages <- false
				return
			}
			res, err := policy.ParsedExpr.Evaluate(params)
			if err != nil {
				logger.Errorf("err is %v", err)
				messages <- false
				return
			}
			if res == true {
				messages <- true
			}
			messages <- false
		}(&f.Policies[i])
	}
	// Wait for responses.
	for i := 0; i < len(f.Policies); i++ {
		if <-messages {
			// First true means we can stop waiting
			return true
		}
	}
	return false
}
