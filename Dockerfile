FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27.3-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /opsmind ./cmd/opsmind

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata && addgroup -S opsmind && adduser -S opsmind -G opsmind
WORKDIR /app
COPY --from=backend /opsmind /app/opsmind
COPY --from=web /src/web/dist /app/web/dist
COPY LICENSE THIRD_PARTY_NOTICES.md /app/
COPY licenses /app/licenses
USER opsmind
EXPOSE 8080
ENTRYPOINT ["/app/opsmind"]
CMD ["serve"]
