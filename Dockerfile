########################
### Builder          ###
########################
FROM golang:1.13 as builder

COPY . /go/src/github.com/github/fault
WORKDIR /go/src/github.com/github/fault/test
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/testfaults

########################
### Final            ###
########################
FROM scratch
COPY --from=builder /go/bin/testfaults /go/bin/testfaults

ENTRYPOINT ["/go/bin/testfaults"]
