FROM golang:1.22 AS builder

WORKDIR /app
COPY . /app/

RUN go build -o disposable .

FROM ubuntu:22.04
COPY --from=builder /app/disposable /usr/local/bin/disposable
RUN apt-get update && apt-get install -y ca-certificates

CMD /usr/local/bin/disposable


