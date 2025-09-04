#syntax=docker/dockerfile:1.18.0-labs@sha256:79cdc14e1c220efb546ad14a8ebc816e3277cd72d27195ced5bebdd226dd1025

FROM golang:1.25.0@sha256:5502b0e56fca23feba76dbc5387ba59c593c02ccc2f0f7355871ea9a0852cebe AS build

WORKDIR /go/src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY --parents ./pkg ./main.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/spannerbackup -trimpath -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9f88e0d99c1ceedbce565fad7d3f96744d15e6919c19c7dafe84a6dd9a80c61

COPY --from=build /go/bin/spannerbackup /spannerbackup

# Entrypoint
ENTRYPOINT ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
