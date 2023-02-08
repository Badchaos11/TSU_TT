FROM golang:1.19-alpine3.16 AS builder

ENV DEPLOY=docker

ADD . /src/app
WORKDIR /src/app
RUN go mod download

RUN go build -o main ./cmd/

FROM alpine:edge
COPY --from=builder /src/app/main /main

EXPOSE 3000
EXPOSE 5432

CMD /main 
