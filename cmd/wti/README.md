# go-dojo
The scaffolding for the Go Dojo we'll be hosting at the Eng Offsite on 10/2/2019

## Welcome to the 2019 Dev Offsite Go Dojo!

We will be building a CLI tool to fetch a Jira ticket and print some details to Standard Output (the screen).  You might notice in `wti.go` that we have the shell of a file.  We've implemented the basics for running the code.  Only problem: none of the actual behavior has been developed yet.

*This is up to you.*  We've even written the tests for you!  Consider them to be the specs for your functions.  Start with the first test and get it to pass.

Test your implementation by running `go test` from the command line.

Once you've gotten the first test passing, go to the next test.

*IMPORTANT:* you must delete the `t.Skip("")` lines when you are ready to start working on a new test.  We didn't want you to start with ALL the tests yelling at you.

### Hints:
Just like in Khan Academy, we've structured the hints to help you along the way. The last hint in a function will give you the code to make the test pass.  Try to do it without the hints, but don't feel bad if you need to use them.  We're all learning!

### Try it when you are done:
```go run wti.go LP-1000```

### Stretch goals:
What else can you do to make this tool more useful?  Some ideas:
- Better error handling and reporting
- Use environment variables or CLI input for some hard-coded values (username, password, etc)
- Display more information to the user
- Allow searching by keyword
- Use the existing benchmark and find the fastest string concatenation method
- Write more integration tests (but keep them separate from unit tests)

If you complete any of the stretch goals, be sure to add tests!!!

### Some details about the username/password stuff you see in the code:

Support for passwords in JIRA REST API basic authentication is deprecated and will be removed in the future. While the Jira REST API currently accepts your Atlassian account password in basic auth requests, we strongly recommend that you use API tokens instead. We expect that support for passwords will be deprecated in the future and advise that all new integrations be created with API tokens.

##### Create an API token
Per the [Create a JIRA Service Token](https://confluence.atlassian.com/cloud/api-tokens-938839638.html?_ga=2.71928019.1521673145.1568996908-1205092387.1568813216) documentation, create an API token from your Atlassian account:

1. Log in to https://id.atlassian.com/manage/api-tokens.
2. Click Create API token.
3. From the dialog that appears, enter a memorable and concise Label for your token and click Create.
4. Click Copy to clipboard, then paste the token to your script, or elsewhere to save:


<table><tr><td>:bulb: <b>NOTE:</b> For security reasons it isn't possible to view the token after closing the creation dialog; if necessary, create a new token.<br/>
You should store the token securely, just as for any password.
</td></tr></table>

##### Test an API token
A primary use case for API tokens is to allow scripts to access REST APIs for Atlassian Cloud applications using HTTP basic authentication.

Depending on the details of the HTTP library you use, simply replace your password with the token. For example, when using curl, you could do something like this:
```shell script
curl -v -L \
https://khanacademy.atlassian.net/rest/api/2/issue/JIRA-116 \
--user $(whoami)@khanacademy.org:${JIRA_API_TOKEN} | jq '{"key": .key, "summary": .fields.summary, "description": .fields.description}'
```
Note that `$(whoami)@khanacademy.org` here is intended to be the email address for the Atlassian account you're using to create the token

##### Test coverage Reports
```
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
go tool cover -func=coverage.out
```

##### Test Benchmarks
If the test suite contains benchmarks, you can run these with the `--bench` and `--benchmem` flags:

```
go test -v --bench . --benchmem
```
Keep in mind that each reviewer will run benchmarks on a different machine, with different specs, so the results from these benchmark tests may vary.

##### Seperating Integration Tests]
You can build the executable and run a simple integration test if you do:
```shell script
go build
PATH=.:$PATH go test -tags=integration
```
It is important to keep slow and possibly destructive integration tests separated from unit tests. 
 
The tag used in the build comment cannot have a dash, although underscores are allowed. For example, `// +build unit-tests` will not work, whereas `// +build unit_tests` will.

##### Resources for future exploration
+ [Prefer table driven tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
+ [Go Walkthrough: encoding/json package](https://medium.com/go-walkthrough/go-walkthrough-encoding-json-package-9681d1d37a8f)
+ [Lesser known features of go test](https://splice.com/blog/lesser-known-features-go-test/)
+ [Unit Testing HTTP Client in Go](http://hassansin.github.io/Unit-Testing-http-client-in-Go)
+ [Peter Bourgon Best Practices](http://peter.bourgon.org/go-best-practices-2016/)
+ [How I write Go after 7 years](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831)
+ [Standard Package Layout](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1)
+ [Structuring Applications in Go](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091)
+ [Dave Cheney's High Performance Go Workshop Content](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html)
+ [Go database/sql tutorial](http://go-database-sql.org/)
+ [How to work with Postgres in Go
](https://medium.com/avitotech/how-to-work-with-postgres-in-go-bad2dabd13e4)

##### Well-known struct tags

Go offers [struct tags](https://golang.org/ref/spec#Tag). Tags are the backticked strings you sometimes see at the end of structs, which are discoverable via reflection. These enjoy a wide range of use in the standard library in the JSON/XML and other encoding packages. 
```
type User struct {
        Name    string `json:"name"`
        Age     int    `json:"age,omitempty"`
        Zipcode int    `json:"zipcode,string"`
}
```
The json struct tag options include:
+ Renaming the field’s key. A lot of JSON keys are camel cased so it can be important to change the name to match.
+ The omitempty flag can be set which will remove any non-struct fields which have an empty value.
+ The string flag can be used to force a field to encode as a string. For example, forcing an int field to be encoded as a quoted string.

The community welcomed struct tags and has built ORMs, further encodings, flag parsers and much more around them since, especially for these tasks, single-sourcing is beneficial for data structures.

Tag       | Documentation
----------|---------------
xml       | https://godoc.org/encoding/xml
json      | https://godoc.org/encoding/json
asn1      | https://godoc.org/encoding/asn1
reform    | https://godoc.org/gopkg.in/reform.v1
dynamodb  | https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbattribute/#Marshal
bigquery  | https://godoc.org/cloud.google.com/go/bigquery
datastore | https://godoc.org/cloud.google.com/go/datastore
spanner   | https://godoc.org/cloud.google.com/go/spanner
gorm      | https://godoc.org/github.com/jinzhu/gorm
yaml      | https://godoc.org/gopkg.in/yaml.v2
validate  | https://github.com/go-playground/validator
mapstructure | https://godoc.org/github.com/mitchellh/mapstructure
protobuf  | https://github.com/golang/protobuf
