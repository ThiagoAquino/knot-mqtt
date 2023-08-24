# First stage: build the application
FROM golang:1.18-alpine AS builder

ARG GOARCH
ARG GOARM

# Install build utilities
RUN apk --no-cache add --virtual .build-deps \
    bash \
    make \
    git \
    curl

# Instale o OpenSSL para gerar o certificado
RUN apk add --no-cache openssl

# Set the Current Working Directory inside the container
WORKDIR /go/src/github.com/CESARBR/knot-mqtt

# Copy the source code from the current directory to $WORKDIR (inside the container)
COPY . .

#Install package required by golang "time" on scratch-based images
RUN apk --no-cache add tzdata

# Install dependencies
RUN apk --no-cache add curl gnupg


# Don't use ca-certificates as .build-deps so that we use import certificates
# from builder, in the final image
RUN apk --no-cache add \
    ca-certificates \
    && update-ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /go/src/github.com/CESARBR/knot-mqtt

# Copy the source code from the current directory to $WORKDIR (inside the container)
COPY . .

# Port that must be opened
EXPOSE 1883
EXPOSE 6060

# Install project development tools and dependencies
RUN go install github.com/ahmetb/govvv@latest
RUN go install gopkg.in/yaml.v2

# Build the application
RUN make bin

# Remove build dependencies
RUN apk del .build-deps

# Second stage: create the entrypoint to the application binary generated in the previous stage
FROM scratch

WORKDIR /app

# Copy the configuration files from the build stage
COPY --from=builder /go/src/github.com/CESARBR/knot-mqtt/internal/ ./internal

# Copy SSL CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary file from the build stage
COPY --from=builder /go/src/github.com/CESARBR/knot-mqtt/app-linux-amd64 knot_mqtt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

VOLUME /app/internal/configuration/

ENV TZ=America/Recife
ENV MQTTIP=localhost
ENV MQTTPORT=1883

ENV DEVICE_CONFIG=internal/configuration/device_config.yaml
ENV KNOT_CONFIG=internal/configuration/knot_setup.yaml
ENV MQTT_SETUP=internal/configuration/mqtt_setup.yaml
ENV MQTT_DEVICE_CONFIG=internal/configuration/mqtt_device_config.yaml

ENTRYPOINT ["./knot_mqtt"]
