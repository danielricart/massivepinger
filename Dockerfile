# Stage 1: dependency cache
FROM golang:1.25 AS deps

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: build
FROM golang:1.25 AS build

WORKDIR /build
COPY --from=deps /root/go/pkg/mod /root/go/pkg/mod
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /massivepinger .

# Stage 3: final image
FROM gcr.io/distroless/static:nonroot AS final

ARG VERSION=dev
ARG REVISION=unknown
ARG CREATED=unknown

LABEL org.opencontainers.image.title="massivepinger" \
      org.opencontainers.image.description="Continuous ICMP ping exporter for Prometheus" \
      org.opencontainers.image.source="https://github.com/example/massivepinger" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${CREATED}"

COPY --from=build /massivepinger /massivepinger

USER nonroot:nonroot

EXPOSE 9123

ENTRYPOINT ["/massivepinger"]
