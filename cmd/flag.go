package main

import (
	"strings"
)

// Patterns is the slice of string.
type Patterns []string 

// Set append string to patterns.
func (p *Patterns) Set(arg string) error {
	*p = append(*p, arg)
	return nil
}

func (p *Patterns) String() (ret string) {
	return strings.Join(*p, ",")
}