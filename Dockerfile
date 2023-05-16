FROM golang
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0"
LABEL maintainer="Garri Gilmanov<garrigilmanov@gmail.com>"
WORKDIR /root/
COPY --from=0 /go/bin/pipeline .
ENTRYPOINT ./pipeline
EXPOSE 8080
