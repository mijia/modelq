ModelQ
===============

ModelQ is a code generator for creating Golang codes/models to access RDBMS database/tables (only MySQL/PostgresQL supported for now).

Updates
---------------

1. Add the template support for generated codes, e.g. examples/custom.tmpl, you can define you own code segments for each part and don't need to define all of them, the generated code is listed in examples/custom/*.go

Simple Idea
---------------

Read the schema from MySQL/PostgresQL database (the whole database or only some tables), and Bang! The go models are there. I embrace the "SQL First" for modeling the business, then use the ModelQ to generate corresponding models for accessing. ModelQ is concerning about two aspects:

1. Easy CRUD interface and query builder but without the golang reflection involved.
2. Facilitate the Go compiler for the correctness (I think this is very important.)

A simple example could be found under `./examples`, it is about a blog model contains Users and Articles. So the database can be set up using the examples/blog.mysql.sql or examples/blog.pq.sql, then run

```
$ modelq -db="root@/blog" -pkg=mysql -tables=user,article -driver=mysql -schema=blog
```

or

```
$ modelq -db="dbname=blog sslmode=disable" -pkg=postgres -tables=user,article -driver=postgres -schema=public
```

Then the models for User and Article would be generated in the directory of "./examples".

CLI Usage
---------------
```
-db="": Target database source string: e.g. root@tcp(127.0.0.1:3306)/test?charset=utf-8
-dont-touch-timestamp=false: Should touch the datetime fields with default value or on update
-driver="mysql": Current supported drivers include mysql, postgres
-p=4: Parallell running for code generator
-pkg="": Go source code package for generated models
-schema="": Schema for postgresql, database name for mysql
-tables="": You may specify which tables the models need to be created, e.g. "user,article,blog"
-template="": Passing the template to generate code, or use the default one
```

You can embed this CLI command in `go generate` tools

API
---------------

Please check the `./examples/model_test.go` to take a glance. Some basic queries are supported, and to take advantage of the compiler, have to use many type guarded funcs defined for the model, e.g.

```
objs := models.UserObjs
users, err := objs.Select("Id", "Name", "Age").
                   Where(objs.FilterAge(">=", 15).Or(objs.FilterAge("IN", 8, 9, 10))).
                   OrderBy("-Age", "Name").Limit(1, 20).List(db)

```

The `Age` of `User` model is a `int`, so go compiler will complain if a `string` is sent in like `objs.FilterAge(">", "15")`. ModelQ will generate all the filters for each field/column of each model then the type requirements would be in the func signatures.

To support different drivers, modelq have to use `gmq.Open` and `gmq.Beginx` for `gmq.Db` and `gmq.Tx` objects, like

```
db, err := gmq.Open("postgres", "dbname=blog sslmode=disable")
tx, err := db.Beginx()
gmq.WithinTx(db, func(tx *gmq.Tx) error {...})
```

Can't do so far
---------------

This is only a early rough implementation, missing a lot of things so far.

* distinct, sum, average and etc. Definitely will get those.
* Joins and Unions. Those seems very likely to the count/distinct/sum and etc. Complicated data structure may be needed.
* The generated models rely on the modelq/gmq package, I am not sure if this would be OK, or could this be changable and plugable, no idea so far.
* No relations for complicated modeling (maybe will never consider this)
* Only MySQL, PostgresQL supported
* Seems github.com/lib/pq has problems to support time.Time scan

But I just want to release it early and get the feedbacks early. So ideas and pull requests would be really welcomed and appreciated!

License
-----
MIT
