# rest-api-go
REST API using the Go programming language.

#### To Start
```
$ go run main.go
```

#### Endpoints
- GET `/quotes` -> All quotes in the quotes document (database)
- GET `/quote/{id}` -> Get a single quote
- POST `/quote/{id}` -> Create a new quote
- DELETE `/quote/{id}` -> Delete a quote

#### References
##### API Application in Go
- [Building a RESTful API with Golang](https://www.codementor.io/codehakase/building-a-restful-api-with-golang-a6yivzqdo)
- [API foundations in Go](https://leanpub.com/api-foundations)

##### Go Development
- [Your First Program](https://www.golang-book.com/books/intro/2)
- [Dep](https://github.com/golang/dep): dependency management tool for Go
- [Godoc: documenting Go code](https://blog.golang.org/godoc-documenting-go-code)

##### Packages
- [time](https://golang.org/pkg/time/)
- [encoding/json](https://golang.org/pkg/encoding/json/)
