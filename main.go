package main

import (
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mijia/modelq/gmq"
	"github.com/mijia/modelq/meta"
)

func main() {
	var targetDb, tableNames, packageName string
	var touchTimestamp bool
	var pCount int
	flag.StringVar(&targetDb, "db", "", "Target database source string: e.g. root@tcp(127.0.0.1:3306)/test?charset=utf-8")
	flag.StringVar(&tableNames, "tables", "", "You may specify which tables the models need to be created, e.g. \"user,article,blog\"")
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

	cfg.dsn = replaceDbName(cfg.dsn, cfg.dbname, "information_schema")
	dbSchema, err := loadTablesMeta(cfg, tableNames)
	if err != nil {
		log.Println("Cannot load table schemas from db.")
		log.Fatalln(err)
	}

	codeConfig := CodeConfig{packageName, touchTimestamp}
	generateModels(cfg.dbname, dbSchema, codeConfig)
	formatCodes(packageName)
}

func formatCodes(pkg string) {
	log.Println("Running gofmt *.go")
	var out bytes.Buffer
	cmd := exec.Command("gofmt", "-w", "-s", pkg)
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Println(out.String())
		log.Fatalf("Fail to run gofmt package, %s", err)
	}
}

func printUsages(message ...interface{}) {
	for _, x := range message {
		fmt.Println(x)
	}
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}

type TableSchema []meta.Columns
type DbSchema map[string]TableSchema

func loadTablesMeta(cfg *_DsnConfig, tableNames string) (DbSchema, error) {
	tables := strings.Split(tableNames, ",")
	log.Printf("Start to load tables schema from db, %s, tables=%s", cfg.dbname, tables)
	db, err := sql.Open("mysql", cfg.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	dbSchema := make(DbSchema)
	if len(tableNames) == 0 {
		err = queryColumns(db, cfg.dbname, nil, dbSchema)
	} else {
		err = queryColumns(db, cfg.dbname, tables, dbSchema)
	}
	if err != nil {
		return nil, err
	}

	log.Printf("Loaded schema data of %d table(s) from db[%s]", len(dbSchema), cfg.dbname)
	return dbSchema, nil
}

func queryColumns(db *sql.DB, dbName string, tables []string, dbSchema DbSchema) error {
	objs := meta.ColumnsObjs
	filter := objs.FilterTableSchema("=", dbName)
	if len(tables) > 0 {
		filter = filter.And(objs.FilterTableName("IN", tables[0], tables[1:]...))
	}

	query := objs.Select().Where(filter).OrderBy("TableName", "OrdinalPosition")
	err := query.Iterate(db, func(col meta.Columns) bool {
		if _, ok := dbSchema[col.TableName]; !ok {
			dbSchema[col.TableName] = make(TableSchema, 0)
		}
		dbSchema[col.TableName] = append(dbSchema[col.TableName], col)
		return true
	})
	return err
}

var (
	errInvalidDSNUnescaped = errors.New("Invalid DSN: Did you forget to escape a param value?")
	errInvalidDSNAddr      = errors.New("Invalid DSN: Network Address not terminated (missing closing brace)")
	errInvalidDSNNoSlash   = errors.New("Invalid DSN: Missing the slash separating the database name")
)

type _DsnConfig struct {
	dsn    string
	user   string
	passwd string
	net    string
	addr   string
	dbname string
}

func replaceDbName(dsn string, oldName, newName string) string {
	return strings.Replace(dsn, "/"+oldName, "/"+newName, 1)
}

// Code taken from github.com/go-sql-driver/mysql
func parseDsn(dsn string) (cfg *_DsnConfig, err error) {
	// New config with some default values
	cfg = &_DsnConfig{dsn: dsn}

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			foundSlash = true
			var j, k int
			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								cfg.passwd = dsn[k+1 : j]
								break
							}
						}
						cfg.user = dsn[:k]
						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return nil, errInvalidDSNUnescaped
							}
							return nil, errInvalidDSNAddr
						}
						cfg.addr = dsn[k+1 : i-1]
						break
					}
				}
				cfg.net = dsn[j+1 : k]
			}

			// dbname[?param1=value1&...&paramN=valueN]
			// Find the first '?' in dsn[i+1:]
			for j = i + 1; j < len(dsn); j++ {
				if dsn[j] == '?' {
					break
				}
			}
			cfg.dbname = dsn[i+1 : j]
			break
		}
	}

	if !foundSlash && len(dsn) > 0 {
		return nil, errInvalidDSNNoSlash
	}

	// Set default network if empty
	if cfg.net == "" {
		cfg.net = "tcp"
	}

	// Set default address if empty
	if cfg.addr == "" {
		switch cfg.net {
		case "tcp":
			cfg.addr = "127.0.0.1:3306"
		case "unix":
			cfg.addr = "/tmp/mysql.sock"
		default:
			return nil, errors.New("Default addr for network '" + cfg.net + "' unknown")
		}
	}

	return
}

func init() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	gmq.Debug = false
}
