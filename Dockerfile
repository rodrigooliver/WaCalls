FROM node:22-alpine AS client
WORKDIR /src/client
COPY client/package.json client/package-lock.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=client /src/client/dist /src/client/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wacalls ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && mkdir -p /data
COPY --from=build /out/wacalls /usr/local/bin/wacalls
COPY --from=build /src/client/dist /client/dist
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["wacalls"]
CMD ["-addr", ":8080", "-db", "/data/wacalls.db", "-static", "/client/dist"]
