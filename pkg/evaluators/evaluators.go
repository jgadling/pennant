package evaluators

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/Knetic/govaluate"
)

// Scale any value to an int from 1 - 100
func libraryPct(arguments ...interface{}) (interface{}, error) {
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
	return float64((num % 100) + 1), nil
}

// Configure the govaluate library to add the pct method to the evaluator
func GetLibraryFunctions() map[string]govaluate.ExpressionFunction {
	functions := make(map[string]govaluate.ExpressionFunction)
	functions["pct"] = libraryPct
	return functions
}
