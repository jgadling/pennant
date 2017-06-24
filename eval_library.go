package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"hash/fnv"
	"math"
)

// Scale any value to an int from 1 - 100
func library_pct(arguments ...interface{}) (interface{}, error) {
	firstArg := arguments[0]

	var value string
	switch firstArg.(type) {
	case float64:
		value = fmt.Sprintf("%f", firstArg)
	default:
		value = fmt.Sprintf("%v", firstArg)
	}

	h := fnv.New64a()
	h.Write([]byte(value))
	num := uint64(math.Mod(float64(h.Sum64()), 100.0))
	logger.Debugf("string value is %v", value)
	logger.Debugf("scaled value is %v", num)
	return float64((num % 100) + 1), nil
}

// Configure the govaluate library to add the pct method to the evaluator
func getLibraryFunctions() map[string]govaluate.ExpressionFunction {
	functions := make(map[string]govaluate.ExpressionFunction)
	functions["pct"] = library_pct
	return functions
}
