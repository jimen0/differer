FROM golang:1.14

ENV GOPATH /go
ENV GOBIN ${GOPATH}/bin
ENV PATH ${PATH}:${GOBIN}

ENV GOOS linux
ENV GOARCH amd64
ENV APP differer
ENV PROJECT github.com/jimen0/${APP}

# install protoc and the plugin for Go
RUN apt-get update && apt-get -y --allow-unauthenticated install unzip wget && apt-get clean
RUN mkdir -p /tmp/protoc \
    && wget -c https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip -O /tmp/protoc/protoc.zip \
    && unzip /tmp/protoc/protoc.zip -d protoc \
    && cp protoc/bin/* /usr/local/bin \
    && cp -R protoc/include/* /usr/local/include \
    && chmod u+x /usr/local/bin/protoc \
    && rm -fr /tmp/protoc protoc

RUN go get -u -v google.golang.org/protobuf/cmd/protoc-gen-go

RUN mkdir -p /go/src/${PROJECT}
ADD . /go/src/${PROJECT}
WORKDIR /go/src/${PROJECT}

RUN protoc --go_out=./scheduler/ ./scheduler/scheduler.proto

RUN CGO_ENABLED=0 go build -mod=readonly \
    -a -tags netgo -installsuffix netgo \
    -ldflags "-s -w -extldflags '-static'" \
    -o /${APP} ${PROJECT}/cmd/${APP}

FROM gcr.io/distroless/static

ENV APP differer
ENV PROJECT github.com/jimen0/${APP}
ENV PORT 8080
ENV DIFFERER_CONFIG /config.yaml
ARG CONFIG_FILE=config.yaml

COPY --from=0 /${APP} /${APP}
COPY --from=0 /go/src/${PROJECT}/${CONFIG_FILE} /config.yaml
EXPOSE ${PORT}

ENTRYPOINT [ "/differer" ]
