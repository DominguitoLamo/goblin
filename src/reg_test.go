package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestReg(t *testing.T) {
    re := regexp.MustCompile(`(?P<variable>[a-zA-Z_][a-zA-Z0-9_]*) = (?P<value>.*)`)
    text := "name = John Doe\nage = 30"

    matches := re.FindStringSubmatch(text)
    if len(matches) > 0 {
        for i, name := range re.SubexpNames() {
            if i != 0 && name != "" { // Skip the entire match
                fmt.Printf("%s: %s\n", name, matches[i])
            }
        }
    }
}