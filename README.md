# Chaos Fault Injection

TODO:

- Maybe get rid of DROP fault type?
- Provide docker image to make requuests against
- GitHub Actions to run CI
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
