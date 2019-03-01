# Be Factory
Be Factory produces go code for a rest api server. It's uses a interactive command line interface to ask the user to describe the SQL table and generates the go struct decribing it, the functions to access to the database, the functions that handle the requests and the test functions for unit testing.

## Structure of the generated code

Be Factory has been designed to generate code according to a certain structure i.e. :

* all tables can be accessed through go structs members located in the models package,

* all requests are handles by functions located in the actions package,

* unit tests are provided by an unique function that calls all the other test functions in the right order to avoid bord effects. Unit tests directly uses requests in order to check routes, authorization, request handling function, models functions and underlying SQL queries.

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
