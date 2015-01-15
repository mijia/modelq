package gmq

import (
	"fmt"
)

func dbQuote(name string, driverName string) string {
	if driverName == "postgres" {
		return fmt.Sprintf("\"%s\"", name)
	}
	return fmt.Sprintf("`%s`", name)
}

func tableNamewithAlias(name, alias, driverName string) string {
	n := dbQuote(name, driverName)
	if alias != "" {
		n = fmt.Sprintf("%s AS %s", n, dbQuote(alias, driverName))
	}
	return n
}

func nameWithAlias(name, alias, driverName string) string {
	n := dbQuote(name, driverName)
	if alias != "" {
		n = fmt.Sprintf("%s.%s", dbQuote(alias, driverName), n)
	}
	return n
}

func genQMarks(count int) string {
	if count == 0 {
		return ""
	}
	marks := make([]byte, count*3-2)
	for i := 0; i < count-1; i++ {
		marks[3*i] = '?'
		marks[3*i+1] = ','
		marks[3*i+2] = ' '
	}
	marks[len(marks)-1] = '?'
	return string(marks)
}
