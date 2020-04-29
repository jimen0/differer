FROM golang:1.14 as build

WORKDIR /go/src/app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -v -o /app .

FROM gcr.io/distroless/base
COPY --from=build /app /app
CMD ["/app"]
