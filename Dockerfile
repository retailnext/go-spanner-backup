#syntax=docker/dockerfile:1.23.0-labs@sha256:7eca9451d94f9b8ad22e44988b92d595d3e4d65163794237949a8c3413fbed5d

FROM golang:1.26.2@sha256:b54cbf583d390341599d7bcbc062425c081105cc5ef6d170ced98ef9d047c716 AS build

WORKDIR /go/src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY --parents ./pkg ./main.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/spannerbackup -trimpath -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

COPY --from=build /go/bin/spannerbackup /spannerbackup

# Entrypoint
ENTRYPOINT ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
