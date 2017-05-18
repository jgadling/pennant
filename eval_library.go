package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"hash/fnv"
	"math"
)

// Scale any value to an int from 1 - 100
func library_pct(arguments ...interface{}) (interface{}, error) {
	// TODO, this is naive and doesn't handle large ints/floats
	// well, but it will get us off the ground for the moment.
	value := fmt.Sprintf("%v", arguments[0])
	h := fnv.New64a()
	h.Write([]byte(value))
	num := uint64(math.Mod(float64(h.Sum64()), 100.0))
	logger.Infof("value is %v", value)
	logger.Infof("hash is %f", float64(h.Sum64()))
	logger.Infof("num is %v", num)
	return float64((num % 100) + 1), nil
}

func getLibraryFunctions() map[string]govaluate.ExpressionFunction {
	functions := make(map[string]govaluate.ExpressionFunction)
	functions["pct"] = library_pct
	return functions
}
