# Toolbox

Random command line tools I use day to day. Currently porting from various shell, Python, and JavaScript scripts to Go for
portability.

One-off, throw-away scripts I tend to run wherever like this:

```
///usr/bin/env go run "$0" "$@" ; exit "$?"

package main

func main() {
    println("Hello world!")
}
```
The things that get moved here are something I use regularly, and __*don't*__ want to always re-compile, and __*do*__ want to revision control.

## Tools


### JP - JIRA Pull
Create a github pull request populated with data from Jira ticket

```
$ jp
```
If you are on a topic branch that matches a jira ticket, and you have committed and pushed your changes, this will convert the jira ticket's Description from Atlassian WikiMarkUp syntax to Github Markdown, and open a new Github pull request with that title and description.

Equivalent to:
```
$ wti $(git rev-parse --abbrev-ref HEAD) -resolves | j2m | gh-make-pull
```

### JIRA to Markdown (J2M)
```
cat ./cmd/j2m/j2m.jira | j2m
```

### What the Issue? (Lookup JIRA ticket)
```
$ wti CORE-5339
```
### Github Make Personal Access Token
```
$ github-make-token "My token for triaging pull requests"
```

### Github Make Pull Request
```
$ github-make-pull "My title for pull requests"
```

### Eureka Lookup: Lookup URL for an instance using Eureka Service discovery
```
$ eureka-lookup search3 prod
```

### Generate TLS Cert

```
$ generate-tls-cert --host=localhost,127.0.0.1
```


