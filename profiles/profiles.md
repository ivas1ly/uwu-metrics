# How to profile with pprof

https://pkg.go.dev/net/http/pprof

## Server

Start the server:

```shell
go run ./cmd/server/main.go -a ":8080" \
-r=true \
-d="postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable" \
-i=10 \
-k="cde0eae6211b525aa851d670d236f6d3c1c29ae74915cb9642c79b38abfccd3c"
```

Run the load using the command from the `loadtesting` directory:

```sh
vegeta attack -duration=40s -rate=10 -targets=target.list
```

Save the pprof heap profile:

```shell
curl "http://127.0.0.1:9090/debug/pprof/heap?seconds=30" > profiles/base_server.pprof
```

Analyze `alloc_space` and `inuse_space`:

```shell
go tool pprof -http=":9095" profiles/base_server.pprof
```

## Agent

Start the server:

```shell
go run ./cmd/server/main.go -a ":8080" \
-r=true \
-d="postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable" \
-i=10 \
-k="cde0eae6211b525aa851d670d236f6d3c1c29ae74915cb9642c79b38abfccd3c"
```

And then start the agent:

```shell
go run ./cmd/agent/main.go -a=localhost:8080 \
-r=4 \
-p=2 \
-k="cde0eae6211b525aa851d670d236f6d3c1c29ae74915cb9642c79b38abfccd3c"
```

Save the pprof heap profile:

```shell
curl "http://127.0.0.1:9091/debug/pprof/heap?seconds=30" > profiles/base_agent.pprof
```

Analyze `alloc_space` and `inuse_space`:

```shell
go tool pprof -http=":9096" profiles/base_agent.pprof
```