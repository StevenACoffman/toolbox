# Golang Tools

This subdirectory is intended to segregate build tools (and their
dependencies) from runtime dependencies and one another. This
will be done by having each tool have it's own
dependency holder subdirectory and go module.

Binaries should be installed without any extra tools (except shell),
just by running `webapp/_tools/install.sh` or a make file equivalent.
The `install.sh` script just runs `ls-imports` utility in the context
of `_tools` module and sends its output to go install command.
The tool `ls-imports` is contained inside `_tools/ls-imports`
directory and all that it does just prints all imports from provided
files.

### Adding new tools

If you have a new tool like `stringer`, which would normally be
installed with:
`go get "golang.org/x/tools/cmd/stringer"`, then assuming `TOOLNAME`
is set to `stringer` and `TOOLGOGET` is set to
`golang.org/x/tools/cmd/stringer`:


1. Add a new subdirectory `mkdir ${TOOLNAME}`
2. `cd ${TOOLNAME}`
3. initialize a new go module there:
`go mod init github.com/Khan/webapp/_tools/${TOOLNAME}`.
4. Run the following shell code:
    ```shell
    cat <<EOD > tools.go
    // +build tools
    
    package tools
    
    import (
            _ "${TOOLGOGET}"
    )
    EOD
    ```
5. By default, the next invocation of go install will update `go.mod` with the
current latest.If you wish to pin to a specific version other than
current latest then specify that version in the go.mod file or by
running:
    ```shell
    go mod edit -require ${TOOLGOGET}@v0.0.0-20200109205111-0235f80b2ba3
    ```
### Background

Currently, our monorepo uses a single go.mod at the very root
(see [ADR #177](https://docs.google.com/document/d/1XcSCeKjfmsyx4sZF-fReMOOB-l94JwSK3WjAT9DVLVk/edit?usp=sharing)).
We need to be cautious of adding new runtime dependencies for several
reasons, but the context of build time tools changes these
considerations.
Certain licenses preclude runtime usage of libraries in private code
bases, but allow their use in open source tools that can then be used
in private code bases.

Build tools can also evolve at a different cadence and with a
different level of risk than our runtime dependencies. If a build
breaks or even if a build succeeds and deploys a new version that
fails in production, or if our code changes at the same time that a
build tool version changes, we cannot determine which of those changes
caused the problem. Pinning build tools to ensure deterministic builds
is considered a best practice.

Before Go 1.13 it was also considered a best practice in the Go
community to add tools as explicit dependencies in a `tools.go`
(https://github.com/golang/go/issues/25922) file. Now with go 1.13.x
`go mod tidy` adds such dependencies to the module's `go.mod` and
`go.sum` files, which we may want to avoid for our root.

See [ADR-300](https://docs.google.com/document/d/1rYk90y3Q4-xaTDomMlP4w5UFteInfrkV2xBomg-rsWU/edit?usp=sharing)
for more context and alternatives considered.
