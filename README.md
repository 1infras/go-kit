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
s := NewServer("test_server", onClose)

routes := transport.Transport{
    PathPrefix: "/api/v1",
    Routes: []transport.Route{
        {
            Path: "/",
            Method: "GET",
            Handler: &ExampleHandler{
                Foo: "bar"
            }
        },
    },
}
  
s.AddRoutes(routes}

s.Run()
}
```

Run with:
```shell
go run main.go -http-port=8080 -log-level=debug -skip-config=true
```

## Contributors
[![](https://github.com/ducmeit1.png?size=50)](https://github.com/ducmeit1)

## Licensed

This repository belongs to 1Infras project of 1MG (One Mount Group)
