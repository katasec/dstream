ARG  GO_VERSION=1.23.2
#----------------------------------------------------------------
# Build stage
#----------------------------------------------------------------

FROM golang:${GO_VERSION} as build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify && go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/dstream


#----------------------------------------------------------------
# Runtime stage
#   with distroless, expected ~ 22MB size of image
#   :debug tag will included shell for debugging purpose
#----------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:debug AS release-stage
WORKDIR /app

COPY --from=build-stage /app/dstream /app/dstream

CMD ["/app/dstream"]

