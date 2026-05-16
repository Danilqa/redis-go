package resp

import (
	"fmt"
	"strings"
)

func BulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func SimpleError(str string) string {
	return fmt.Sprintf("-%s\r\n", str)
}

func NullBulkString() string {
	return "$-1\r\n"
}

func NullArray() string {
	return "*-1\r\n"
}

func Integer(num int) string {
	return fmt.Sprintf(":%d\r\n", num)
}

func Array(values []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(values))
	for _, v := range values {
		b.WriteString(BulkString(v))
	}
	return b.String()
}

func RawArray(elements []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(elements))
	for _, e := range elements {
		b.WriteString(e)
	}
	return b.String()
}
