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
    - Viper (Read application configuration)

## Usages

### Go Module Private

If you meet any error when fetching this repository, please add

```shell
export GOPRIVATE="gitlab.id.vin/devops"
```

Set git config:

```shell
git config --global url."git@gitlab.id.vin:".insteadOf "https://gitlab.id.vin"
```

In case you haven't allow to do set git config, you could add these line into your `.gitconfig`:

```shell
vim ~/.gitconfig
#Add these lines to the end
[url "git@gitlab.id.vin:"]
insteadOf = https://gitlab.id.vin/

#Save with ESC and :wq!
```

Example to download this module:

```shell
go get -u gitlab.id.vin/devops/go-kit@v0.1.1
```
