package server

import (
	"fmt"
	"strings"
)

type Info map[string]InfoCategory

type InfoCategory map[string]string

func (info *Info) Get(key string) string {
	lines := []string{}
	for key := range *info {
		category := (*info)[key]
		for key := range category {
			lines = append(lines, fmt.Sprintf("%s:%s", key, category[key]))
		}
	}
	return strings.Join(lines, "")
}
