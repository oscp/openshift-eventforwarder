FROM golang:1.7-wheezy

WORKDIR /go/src/github.com/oscp/ose-event-forwarder

COPY . /go/src/github.com/oscp/ose-event-forwarder

RUN go get golang.org/x/build/kubernetes/api

RUN go install -v

CMD ["ose-event-forwarder"]