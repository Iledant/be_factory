# Be Factory
Be Factory produces go code for a rest api server. It uses an interactive command line interface to ask the user to describe the SQL table and which actions should be codded and generates the go struct describing it, the functions to access to the database, the functions that handle the requests and the test functions for unit testing.

## Pattern of the generated code

Be Factory has been designed to generate code according to a certain pattern i.e. :

* all tables can be accessed through go structs members located in the models package and directory. These members use direct lib/pq queries to be efficient and specially dedicated,

* all requests are handles by functions located in the actions package using models to store or fetch datas according to a MVC inspired model,

* all routes are located in one file using middlewares to decrypt json web tokens, check if the user is connected and if the user has the correct rights to access to a route,

* unit tests are provided by an unique function located in `common_test.go` that calls all the other dedicated test functions in the right order to avoid border effects.

Unit tests directly use http requests in order to check routes, authorization, request handling functions, models functions and underlying SQL queries.

The structure is therefore the following one

```
main.go
  /models
    table1.go
    table2.go
    ...
  /actions
    commons.go
    commons_test.go
    routes.go
    table1.go
    table1_test.go
    table2.go
    table2_test.go
    ...
```

`common_test.go` is used to configure the test database, deleting tables created in the previous unit test sequence and creating new ones. Table are empty and unit tests sequences is designed to create a new row, update it, get all rows and finally to delete it.

Be Factory scan `routes.go`, `common_test.go` in order to find good location to insert the generated code i.e. new routes, test table deleting and creating, test functions calls.

Be Factory checks if `/models` and `/routes` directories exist, tries to create them and insert files according to the given name.

## Table description

Be Factory asks the user informations i.e.

* Name of the model. That name is used as go type in the model file

* French name. It is used in the error messages sent backs par action functions

* Fields descriptions i.e.

  * Name. It's must be the go name of the structure and therefore camel cased. If the user types ID Be Factory considers it an unique int ID

  * Type. Be Factory proposes choice between these SQL types : bigint, int, boolean, varchar, text, double precision, date. If varchar is chosen Be Factory asks for the max length

  * Nullable. For SQL table, a non nullable field will be created with the NOT NULL constraint. For go struct, a nullable field will be codded with a special type defined in `commons.go` for example NullInt64

* Actions that must be codded. The possible choices are create, update, get all, get, delete and batch

The different naming convention is the following one :

* Name of the model must be capitalized for example FinancialCommitment. The model name is used for function names for example CreateFinancialCommitment. Be Factory use the `title` go strings package function to ensure that the first letter of the model name is capitalized

* Name of the SQL table must be kebab cased for example financial_commitment. The table name is also used for the file name for example financial_commitment.go in the action and models folders


## Known bugs

Be Factory is used to generate a quick and almost finalized code for REST API. Some TODO comments are inserted in the generated code for the parts that must be checked or completed, for example validation or some unit tests.

Be Factory don't check coherence between chosen actions : for example if the create action is not used, the delete test will be codded as if the ID of the previously created item is available.