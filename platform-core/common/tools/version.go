package tools

import (
	"math"
	"strconv"
	"strings"
)

func CompareVersion(v1, v2 string) bool {
	return parseVersionToNum(v1) > parseVersionToNum(v2)
}

func parseVersionToNum(input string) float64 {
	if input == "" {
		return 0
	}
	input = strings.ToLower(input)
	if strings.HasPrefix(input, "v") {
		input = input[1:]
	}
	var num float64
	for i, v := range strings.Split(input, ".") {
		intV, _ := strconv.Atoi(v)
		num += float64(intV) * math.Pow(100, float64(4-i))
	}
	return num
}
