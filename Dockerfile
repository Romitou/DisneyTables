FROM alpine:3.17 AS go

WORKDIR /app/go
RUN apk update
RUN apk upgrade
RUN apk add go
ADD . .
ENV GOPATH /app
RUN go get
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -o disneytables


FROM alpine:3.17

WORKDIR /app
COPY . .
COPY --from=go /app/go/disneytables /app/disneytables
RUN chmod +x ./disneytables
CMD ["./disneytables"]
