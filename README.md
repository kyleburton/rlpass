# Overview

Wrapper for the `lastpass-cli` tooling implemented in golang.  This wrapper represents a very opinionated way of storing configuration in lastpass so that it is machine readable and consistent.  With a focus on machine readability.



Examples


```
go test *.to
go run main.go ls | jq -C . | less
go run main.go show 6919840240772558827686 | jq -C . | less
```

```
go run mian.go spec
```


Q: Why not just `lpass export --color=never`?

A: Though this allows for bulk export, it does not preserve the fidelity of the information stored in lastpass.

https://github.com/lastpass/lastpass-cli/issues/263
