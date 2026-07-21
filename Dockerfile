#syntax=docker/dockerfile:1.25.0-labs@sha256:4426b5e269e36911b94fb79cf67f1fd7155ef11b2bbc8ab23cbfcbc97130efe9

FROM golang:1.26.5@sha256:3aff6657219a4d9c14e27fb1d8976c49c29fddb70ba835014f477e1c70636647 AS build

WORKDIR /go/src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY --parents ./pkg ./main.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/spannerbackup -trimpath -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639

COPY --from=build /go/bin/spannerbackup /spannerbackup

# Entrypoint
ENTRYPOINT ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
