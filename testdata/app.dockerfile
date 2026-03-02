# Dockerfile with line continuations
FROM golang:1.21-alpine

# Multi-line ENV declaration
ENV APP_NAME=myapp \
    APP_VERSION=1.0.0 \
    APP_PORT=8080

# Multi-line ARG
ARG GOOS=linux \
    GOARCH=amd64

RUN echo "Building $APP_NAME version $APP_VERSION"
RUN echo "Target: ${GOOS}/${GOARCH}"

WORKDIR /app
COPY . .

RUN go build -o main .

CMD ["./main"]
