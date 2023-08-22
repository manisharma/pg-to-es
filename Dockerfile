FROM golang:1.21.0 AS build-stage
ARG directory
ENV GO111MODULE=on
WORKDIR /app

COPY . .
RUN go mod download
COPY .env /app/.env
WORKDIR /app/cmd/${directory}

RUN CGO_ENABLED=0 GOOS=linux go build -o binary

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS release-stage

ARG directory
WORKDIR /

COPY --from=build-stage /app/cmd/${directory}/binary /binary
COPY --from=build-stage /app/.env /.env

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/binary"]