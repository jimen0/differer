## How to create a runner?

This document explains how to build a runner. Go will be used as an example. If you just want to see the code, please visit [this link](./example_runner).

The basic idea behind the runners is that, as long as they listen on HTTP for `Jobs` and return `Results`, the `differer` doesn't care about implementation details behind them. Also, by using the Protocol Buffers messages, the message schema is described only in one place and how each runners processes it only depends on the [`protoc`](https://github.com/protocolbuffers/protobuf) compiler.
