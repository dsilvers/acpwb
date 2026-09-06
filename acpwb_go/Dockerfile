FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /acpwb_go ./cmd/acpwb_go

FROM alpine:3.20
COPY --from=build /acpwb_go /acpwb_go
ENV LISTEN_ADDR=0.0.0.0:8091
EXPOSE 8091
ENTRYPOINT ["/acpwb_go"]
