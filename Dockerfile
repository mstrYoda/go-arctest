FROM golang:1.20-alpine

LABEL maintainer="mstrYoda"
LABEL repository="https://github.com/mstrYoda/go-arctest"
LABEL homepage="https://github.com/mstrYoda/go-arctest"
LABEL "com.github.actions.name"="Go Architecture Test"
LABEL "com.github.actions.description"="Run architecture tests for Go projects"
LABEL "com.github.actions.icon"="check-circle"
LABEL "com.github.actions.color"="green"

WORKDIR /app

# Copy the source code
COPY . .

# Build the CLI tool
RUN go build -o /usr/local/bin/arctest ./cmd/arctest

# Set the entrypoint
ENTRYPOINT ["arctest"]
