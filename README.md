# Chaos Fault Injection

## Docker

```shell
docker build -t faulttest .
docker run -d -p 3000:3000 faulttest
curl -v localhost:3000
```

## Testing

```shell
go test -v -cover -race
```

TODO:

- Readme shows how to launch godoc
- Makefile for tests, docs, docker
- Benchmarks
- Implementation examples
- Implementations for non net/http middlewares?
- Link to initial proposal
- Function to update options on an existing Fault struct
- If I want a guaranteed chain (30% slow and then always reject those slowed) do I need to pass a header of some sort to enable that?

Extend Options

- Modify response body with param?
- Return faults based on headers passed in?
- Set a fault-enabled header on some responses?
- Add a latency distribution instead of just one time
