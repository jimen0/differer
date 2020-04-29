## differer

Differer project aims to help Bug Bounty Hunters to find differences between several languages and libraries URL parsers. Not all of them behave in the same way and it might lead to unexpected vulnerabilities.

URLs format is defined in [RFC 3986](https://tools.ietf.org/html/rfc3986), however there are small differences between languages, libraries and how they deal with incorrect URLs. Some of them report an error to the caller, other raise exceptions and other go with the best-effort approach and try to fix them for you. It is exactly there where unexpected security issues might arise.

```
         foo://example.com:8042/over/there?name=ferret#nose
         \_/   \______________/\_________/ \_________/ \__/
          |           |            |            |        |
       scheme     authority       path        query   fragment
          |   _____________________|__
         / \ /                        \
         urn:example:animal:ferret:nose
```

A lot of work has been done in this particular topic already. One of the most popular places where it has been discussed is on Orange's presentation _[A New Era of SSRF - Exploiting URL Parser in Trending Programming Languages!](https://www.blackhat.com/docs/us-17/thursday/us-17-Tsai-A-New-Era-Of-SSRF-Exploiting-URL-Parser-In-Trending-Programming-Languages.pdf)_.

This project doesn't bring any new attack technique, rather than that it tries to make the process of finding parser differences easier.

### The goal

To be able to run the parsers against the desired URLs on demand and without worrying about setting up the compilers, interpreters or the messaging broker. The main use case of this project is to be deployed under Google Cloud Platform using the smallest amount of resources and letting it tear down the services if they aren't being used.

The only thing I want to do is to submit a URL somewhere and get the different parsers results.

### Show me the numbers

> Please, remember that this project focuses on maintainability, not on performance.

Here are the numbers when running the project with `400` URLs against `3` parsers. Parser instances located in `europe-west1`, have `128 MiB` of RAM, `1` vCPU, max `4` instances per parser and up to `60` concurrent requests for each instance. The parsers are:

| Language | Version       | Parser                                                                                   |
|----------|---------------|------------------------------------------------------------------------------------------|
| Go       | 1.14.2        | [`net/url.Parse`](https://pkg.go.dev/net/url?tab=doc#Parse )                             |
| Python   | 3.8           | [`urllib.urlparse`](https://docs.python.org/3/library/urllib.parse.html)                 |
| Node     | 14            | [`url`](https://nodejs.org/api/url.html)                                                 |


```console
$ curl -o /dev/null -X'POST' -d @data.json -s -w "%{time_total}\n" "https://REDACTED/differer"
0.888706
```

For cold loads (calls that happen after cloudrun shuts down the containers) time is higher, as expected:

```console
$ curl -o /dev/null -X'POST' -d @data.json -s -w "%{time_total}\n" "https://REDACTED/differer"
2.760033
```

### Why microservices?

Because I don't want to run the tools locally each time I want to see how different languages parse an URL. I simply run a query to my service and get the output.

Setting the project up using App Engine and Cloud Run allows me to forget about infrastructure. GCP shuts my services down and up, plus allows me to restrict the access to them thanks to the firewall and IAM rules.

![GCP architecture](./docs/differer.svg)

However, the project can be used locally too. See [local setup docs](./docs/RUN_LOCAL.md) or [GCP docs](./docs/RUN_GCP.md).

### How to configure?

The configuration file is a simple YAML file. Here is an example if you want to run the project locally with Go's, Python's and Node's parsers. See the [`config_example.yaml`](./config_example.yaml) file for a raw example.

```yaml
---
runners:
  golang: http://golang-parseurl:8082/
  python3_urllib_urlparse: http://python-parseurl:8083/
  node_url_parse: http://node-parseurl:8084/
timeout: 10s
```

### How to add a language or library?

As long as your new `runner` listens on HTTP for a `POST` request containing the jobs, the service is agnostic and doesn't care about where or how you run each runner.

The `Job` and `Result` structure can be found into the [Protocol Buffer](./scheduler/scheduler.proto) the project uses. Use the [`protoc`](https://github.com/protocolbuffers/protobuf) compiler to generate your language's jobs and results parsers. See [this document](./docs/RUNNER_EXAMPLE.md) for a complete example using Go.

Once your `runner` is deployed somewhere, just edit your `config.yaml` adding it.

### How to run locally?

For simplicity, let's assume you run all the services using Docker containers. Follow the [local setup guide](./docs/RUN_LOCAL.md) and then just send a request to `differer` with the URLs you want it to parse.

<details>
<summary>Local run</summary>
<br>

```bash
$ curl -s --request POST 'http://127.0.0.1:8080/differer' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "addresses": [
        "https://google.com:443/foobar",
        "http://user:legit.com@attacker.com/?pwnz=1"
    ]
}' | jq .
{
  "results": [
    {
      "runner": "python3_urllib_urlparse",
      "string": "https://google.com:443/foobar",
      "outputs": {
        "id": "python3:urllib:urlparse",
        "value": "Scheme=https; Host=google.com:443; Path=/foobar;"
      }
    },
    {
      "runner": "python3_urllib_urlparse",
      "string": "http://user:legit.com@attacker.com/?pwnz=1",
      "outputs": {
        "id": "python3:urllib:urlparse",
        "value": "Scheme=http; Host=user:legit.com@attacker.com; Path=/; User=user:legit.com;"
      }
    },
    {
      "runner": "node_url_parse",
      "string": "http://user:legit.com@attacker.com/?pwnz=1",
      "outputs": {
        "id": "node14:url.parse",
        "value": "Scheme=http:; Host=attacker.com; Path=/; User=user:legit.com"
      }
    },
    {
      "runner": "node_url_parse",
      "string": "https://google.com:443/foobar",
      "outputs": {
        "id": "node14:url.parse",
        "value": "Scheme=https:; Host=google.com:443; Path=/foobar;"
      }
    },
    {
      "runner": "golang",
      "string": "https://google.com:443/foobar",
      "outputs": {
        "id": "golang",
        "value": "Scheme=https; Host=google.com:443; Path=/foobar;"
      }
    },
    {
      "runner": "golang",
      "string": "http://user:legit.com@attacker.com/?pwnz=1",
      "outputs": {
        "id": "golang",
        "value": "Scheme=http; Host=attacker.com; Path=/; User=user:legit.com;"
      }
    }
  ]
}
```
</details>

### Why do the runners only accept one task at a time?

Indeed the service would be faster if runners would accept multiple tasks at a time, and changing it to support them would be straight forward. However, I decided to keep it as simple as possible as it's performant enough for me.

### It would be faster if it used `xyz`

The project aims for maintainability and ease of use over performance. Feel free to fork it if you disagree.

### How can I contribute?

Please, check the [contributing documentation](./docs/CONTRIBUTE.md).

### Credits

I decided to build this project after a discussion with some friends ([Karel](https://github.com/karelorigin), [Karim](https://github.com/KarimPwnz), [Corben](https://github.com/lc) and [Amal](https://github.com/amalmurali47)).
