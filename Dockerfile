#syntax=docker/dockerfile:1.27.0-labs@sha256:ae9cc40df4eb5b6adcac0a49bdd8e43b6d29d81087fefae2ceb6fe248aab24c8

FROM golang:1.27.1@sha256:1cfcdb11f37fce9429f617100f39e0251748bbaba454bd431275155701765058 AS build

WORKDIR /go/src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY --parents ./pkg ./main.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/spannerbackup -trimpath -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

COPY --from=build /go/bin/spannerbackup /spannerbackup

# Entrypoint
ENTRYPOINT ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
