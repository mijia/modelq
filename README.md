ModelQ
===============

ModelQ is a code generator for creating Golang codes/models to access RDBMS database/tables (only MySQL supported for now).

Simple Idea
---------------

Read the schema from MySQL database (the whold database or only some tables), and Bang! The go models are there. I embrace the "SQL First" for modeling the business, then use the ModelQ to generate corresponding models for accessing. ModelQ is concerning about two aspects:

1. Easy CRUD interface and query builder but without the golang reflection involved.
2. Facilitate the Go compiler for the correctness (I think this is very important.)

A simple example could be found under `./examples`, it is about a blog model contains Users and Articles. So the database can be set up using the examples/db.sql, then run

```
$ modelq -db="root@/blog" -pkg=models -tables=user,article
```

Then the models for User and Article would be generated in the directory of "./models".

CLI Usage
---------------
```
-db="": Target database source string: e.g. root@tcp(127.0.0.1:3306)/test?charset=utf-8
-dont-touch-timestamp=false: Should touch the datetime fields with default value or on update
-p=4: Parallell running for code generator
-pkg="": Go source code package for generated models
-tables="": You may specify which tables the models need to be created, e.g. "user,article,blog"
```

I haven't tried the `go generate` from go 1.4 yet. Will definitely check it out.

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

Can't do so far
---------------

This is only a early rough implementation, missing a lot of things so far.

* count(*), distinct, sum, average and etc. Definitely will get those, but still need to think it through.
* Joins and Unions. Those seems very likely to the count/distinct/sum and etc. Complicated data structure may be needed.
* The generated models rely on the modelq/gmq package, I am not sure about if this would be OK, or could this be changable and plugable, no idea so far.
* No relations for complicated modeling (maybe will never consider this)
* Only MySQL supported

So ideas and pull requests would be really welcomed and appreciated!