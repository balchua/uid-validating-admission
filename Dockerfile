FROM golang:1.11

ENV GO111MODULE=on
# Copy the code from the host and compile it
WORKDIR $GOPATH/src/uid-vallidating-webhook

COPY go.* ./
COPY main.go ./
COPY server/* server/
COPY config/* config/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix nocgo -o /usr/local/bin/uid-validating-webhook .

ENTRYPOINT ["/usr/local/bin/uid-validating-webhook"]