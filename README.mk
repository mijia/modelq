I am trying to build something called ModelQ which is a go code generator for accessing a RDBMS from Golang.

Simple Idea: 

Just read schema from mysql databases, and Bang! Your go models are there. You can do the CRUD with the model, also the queries (yeah, the queries), but without the relation part like foreign keys, one-to-one or one-to-many.

I am still stucking in the API design phase, so far it can work, but the code is not that beaufiful. But so far you can run after go build:

$ ./modelq -db="root@/blog" -pkg=models -tables=user,article

2014-12-09: 

I just stumbled upon some article (http://www.hydrogen18.com/blog/golang-orms-and-why-im-still-not-using-one.html). In the article, the author also proposed a framework to generate code for RDBMS models since code generation is expected to be standardized in Go 1.4. He thought the features should include:

1. Generate code by parsing the SQL DDL statements of the database.
2. Expose relational concepts in any generated objects.
3. Use a fluent interface to expose a query builder. The resulting SQL should be non-ambiguous.
4. Facilitate greater compile time correctness checking of code.

The 3rd one is difficult, I am still thinking about this. Maybe there exists some algebra data types/interfaces for this query builder.