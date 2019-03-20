###### Go-lang HTTP handler, which automatically creates, registers and updates prometheus metrics.
This is a wrapper around go-lang http handler function, which automatically creates and registers the following prometheus metrics
- total request count
- total error count
- request time (ms) summary: 0.5 0.9 and 0.99 percentile
###### Usage
See main.go for example of usage.
Compile and run main.go
Open another terminal and invoke the following command
curl http://localhost:8081/test
curl http://localhost:8081/test/error
curl http://localhost:8081/metrics | grep test

Continue making requests to /test and /test/error endpoints and observe changes in /metrics endpoint


Contribution guide:
- clone repository
- make changes in a separate branch
- update tests
- create pull request

---
Licensed under Apache 2.0 license
