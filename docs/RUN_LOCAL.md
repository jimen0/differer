## Run the project locally

This guide assumes you have `docker` installed and your version supports multi-stage builds.

- `differer`

    ```console
    $ docker network create --driver bridge differer
    # remember to create your config.yaml file
    $ docker build --build-arg=CONFIG_FILE=<YOUR_CONFIG_FILE.yaml> -t differer .
    $ docker run --rm -p 8080:8080 -e PORT=8080 --network differer differer
    ```

- The runners are independent containers. For example, here is how I set up Python 3 and Go runners.

    ```console
    $ docker run --rm -e PORT=8083 -p 8083:8083 --network differer --name="python-parseurl" gcr.io/REDACTED/python-parseurl:latest
    $ docker run --rm -e PORT=8082 -p 8082:8082 --network differer --name="golang-parseurl" gcr.io/REDACTED/golang-parseurl:latest
    ```

- Then the configuration file can use the container names to find the runners.

    ```yaml
    ---
    runners:
      golang: http://golang-parseurl:8082/
      python3_urllib_urlparse: http://python-parseurl:8083/
    timeout: 10s
    ```
