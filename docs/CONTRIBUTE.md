## Contribute

If you are modifying Go code, please make sure it's as idiomatic as possible and that it satisfies at least `gofmt`, `golint` and `go vet`. Only the last stable release will be supported. There's a `golangci-lint` configuration file [`.golangci.yml`](../.golangci.yml) in case you want to ensure your code is idiomatic.

Docker images changes should aim to keep them small but simple and avoid adding unneeded layers. Only the last stable version of Docker will be supported and images will assume [`buildkit`](https://github.com/moby/buildkit) is enabled.

Please, remember that this is a side project and reviewing proposals or PRs and fixing bugs might take some time.
