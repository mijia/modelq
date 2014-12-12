package gmq

import (
	"fmt"
)

func dbQuote(name string) string {
	return fmt.Sprintf("`%s`", name)
}

func tableNamewithAlias(name, alias string) string {
	n := dbQuote(name)
	if alias != "" {
		n = fmt.Sprintf("%s AS %s", n, dbQuote(alias))
	}
	return n
}

func nameWithAlias(name, alias string) string {
	n := dbQuote(name)
	if alias != "" {
		n = fmt.Sprintf("%s.%s", dbQuote(alias), n)
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
