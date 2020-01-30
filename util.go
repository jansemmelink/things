package main

import "strconv"

func intParam(s string, defaultValue int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return v
}
