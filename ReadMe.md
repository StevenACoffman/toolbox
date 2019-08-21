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

