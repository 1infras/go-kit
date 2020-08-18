# GO-KIT

## Introduction

This repository will help you build a Golang microservices project faster and optimized

- Actor: Golang projects
- Programming language: Golang
- Feature supports:
    - Gorilla Mux (HTTP Router)
    - Negroni (HTTP Middleware)
    - Elastic APM (Elastic Application Performance Monitoring)
    - OpenTracing (Tracing)
    - ElasticSearch (DocumentDB)
    - Redis (Multiple client)
    - Viper (Read application configuration)

## Usages

```shell
go get -u https://github.com/1infras/go-kit
```

- Get starting with:

**main.go**

```golang
type ExampleHandler struct {
    Foo string
}

func (h *ExampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    transport.OKJson(w, map[string]interface{}{
        "foo": h.Foo
    })
}

func onClose() {}

func main() {
    //Create a new HTTP Server with name and close function to callback when closing the server
    s := NewServer("test_server", onClose)

    routes := transport.Transport{
        PathPrefix: "/api/v1", //Set path prefix for every route
        Routes: []transport.Route{
            {
                Path: "/", //Path
                Method: "GET", //HTTP Method
                Handler: &ExampleHandler{ //HTTP Handler
                    Foo: "bar"
                }
            },
        },
    }
    //Adding routes to transport
    s.AddRoutes(routes)
    //Starting run HTTP Server
    s.Run()
}
```

Run with:

```shell
go run main.go -http-port=8080 -log-level=debug -skip-config=true
```

Verify with:

```shell
curl -X GET http://localhost:8080/health
```

or

```shell
curl -X GET http://localhost:8080/api/v1
```

## Contributors

[![ducmeit1](https://github.com/ducmeit1.png?size=50)](https://github.com/ducmeit1)
