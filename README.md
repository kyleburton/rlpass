# Overview

Wrapper for the `lastpass-cli` tooling implemented in golang.  This wrapper represents a very opinionated way of storing configuration in lastpass so that it is machine readable and consistent.  With a focus on machine readability.



Examples


```
go run main.go ls | jq -C . | less
go run main.go show 6919840240772558827686 | jq -C . | less
```
