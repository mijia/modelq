package gmq

import (
	"bytes"
	"fmt"
	"strconv"
)

func dbQuote(name string, driverName string) string {
	if driverName == "postgres" {
		return fmt.Sprintf("\"%s\"", name)
	}
	return fmt.Sprintf("`%s`", name)
}

func tableNamewithAlias(schema, name, alias, driverName string) string {
	n := dbQuote(name, driverName)
	if alias != "" {
		n = fmt.Sprintf("%s AS %s", n, dbQuote(alias, driverName))
	}
	if driverName == "postgres" {
		n = fmt.Sprintf("%s.%s", dbQuote(schema, driverName), n)
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

func paramMarkers(count int, driverName string) string {
	if count == 0 {
		return ""
	}
	marks := make([][]byte, count)
	for i := 1; i <= count; i++ {
		marks[i-1] = []byte("?")
	}
	return string(bytes.Join(marks, []byte(", ")))
}

// code taken from sqlx
func rebindSqlParams(query string, driverName string) string {
	if driverName == "postgres" {
		qb := []byte(query)
		rqb := make([]byte, 0, len(qb)+10)
		j := 1
		for _, b := range qb {
			if b == '?' {
				rqb = append(rqb, '$')
				for _, b := range strconv.Itoa(j) {
					rqb = append(rqb, byte(b))
				}
				j++
			} else {
				rqb = append(rqb, b)
			}
		}
		return string(rqb)
	}
	return query
}
