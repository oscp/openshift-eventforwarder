FROM golang:1.8-alpine

WORKDIR /go/src/github.com/oscp/ose-event-forwarder

COPY . /go/src/github.com/oscp/ose-event-forwarder

RUN go install -v

CMD ["ose-event-forwarder"]