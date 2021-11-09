FROM golang:1.16.3-alpine as builder
ENV GO111MODULE=on
WORKDIR /go/src/github.com/tiniyo/neoms
COPY . .
RUN apk update
RUN apk upgrade
RUN apk add --update gcc g++ libxml2-dev

RUN GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl bash libxml2
WORKDIR /
COPY --from=builder /go/src/github.com/tiniyo/neoms/app .
COPY --from=builder /go/src/github.com/tiniyo/neoms/config.toml /etc/
COPY --from=builder /go/src/github.com/tiniyo/neoms/entrypoint.sh .
COPY --from=builder /go/src/github.com/tiniyo/neoms/TinyMLSchema.xsd .

# Health Check for the service
HEALTHCHECK --timeout=5s --interval=3s --retries=3 CMD curl --fail http://localhost:9092/api/v1/health || exit 1

# Expose the application on port 8080.
# This should be the same as in the app.conf file
EXPOSE 9092

RUN chmod 755 /entrypoint.sh && \
	chown root:root /entrypoint.sh

CMD ["/entrypoint.sh"]
