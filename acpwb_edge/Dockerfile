FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /acpwb_edge ./cmd/acpwb_edge

FROM alpine:3.20
COPY --from=build /acpwb_edge /acpwb_edge
ENTRYPOINT ["/acpwb_edge"]
