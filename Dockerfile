#syntax=docker/dockerfile:1.18.0-labs@sha256:79cdc14e1c220efb546ad14a8ebc816e3277cd72d27195ced5bebdd226dd1025

FROM golang:1.25.1@sha256:b773c94fd926723a415b28e615016a49793764e270566a82c6a47afb4c52f7de AS build

WORKDIR /go/src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY --parents ./pkg ./main.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/spannerbackup -trimpath -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:e8a4044e0b4ae4257efa45fc026c0bc30ad320d43bd4c1a7d5271bd241e386d0

COPY --from=build /go/bin/spannerbackup /spannerbackup

# Entrypoint
ENTRYPOINT ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
