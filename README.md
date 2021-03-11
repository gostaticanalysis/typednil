# typednil

[![pkg.go.dev][gopkg-badge]][gopkg]

`typednil` finds a comparition between typed nil and untyped nil.

```go
func f() {
	err := e()
	if err != nil { // true
		print(err)
	}
}

func e() error {
	var err *struct{error}
	return err
}
```

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/gostaticanalysis/typednil
[gopkg-badge]: https://pkg.go.dev/badge/github.com/gostaticanalysis/typednil?status.svg
