package main

import (
	"fmt"
	"strings"
)

func ToBulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func ToSimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func ToSimpleError(str string) string {
	return fmt.Sprintf("-%s\r\n", str)
}

func ToNullBulkStrings() string {
	return "$-1\r\n"
}

func ToNullArray() string {
	return "*-1\r\n"
}

func ToInteger(num int) string {
	return fmt.Sprintf(":%d\r\n", num)
}

func ToArray(values []string) string {
	return fmt.Sprintf("*%d\r\n%s", len(values), strings.Join(values, ""))
}
