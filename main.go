package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"runtime"
)

func main() {
	var targetDb, tableNames, packageName string
	var cleanPackage, touchTimestamp bool
	var pCount int
	flag.StringVar(&targetDb, "db", "", "Target database source string: e.g. root@tcp(127.0.0.1:3306)/test?charset=utf-8")
	flag.StringVar(&tableNames, "tables", "", "You may specify which tables the models need to be created")
	flag.BoolVar(&cleanPackage, "clean", false, "If clean the generated code directory")
	flag.StringVar(&packageName, "pkg", "", "Go source code package for generated models")
	flag.BoolVar(&touchTimestamp, "dont-touch-timestamp", false, "Should touch the datetime fields with default value or on update")
	flag.IntVar(&pCount, "p", 4, "Parallell running for code generator")
	flag.Parse()

	runtime.GOMAXPROCS(pCount)

	if targetDb == "" {
		fmt.Println("Please provide the target database source.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}
	if packageName == "" {
		printUsages("Please provide the go source code package name for generated models.")
		return
	}

	cfg, err := parseDsn(targetDb)
	if err != nil {
		printUsages("The target database source string doesn't seem correct...", err)
		return
	}
	if cfg.dbname == "" {
		printUsages("Please provide the target database name.")
		return
	}

	dbSchema, err := loadTablesMeta(cfg, tableNames)
	if err != nil {
		log.Println("Cannot load table schemas from db.")
		log.Fatalln(err)
	}
	fmt.Println(dbSchema)
}

func printUsages(message ...interface{}) {
	for _, x := range message {
		fmt.Println(x)
	}
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
