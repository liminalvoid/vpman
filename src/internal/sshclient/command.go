package sshclient

import "strings"

type Command struct {
	Name string
	Args []string
}

func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (c *Command) String() string {
	parts := make([]string, 0, len(c.Args) + 1)
	parts = append(parts, c.Name)
	for _, a := range c.Args {
		parts = append(parts, quote(a))
	}

	return strings.Join(parts, " ")
}
