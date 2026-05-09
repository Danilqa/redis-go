package server

import (
	"fmt"
	"strings"
)

type Info map[string][]string

func (info *Info) Get(key string) string {
	lines := []string{}
	for key := range *info {
		value := (*info)[key]
		lines = append(lines, fmt.Sprintf("%s:%s", value[0], value[1]))
	}
	return strings.Join(lines, "")
}
