FROM golang:latest as build
WORKDIR /src
COPY . .
RUN go mod tidy -v;go build -o /tg-gemini-bot
FROM alpine:latest
WORKDIR /src
COPY --from=build /tg-gemini-bot tg-gemini-bot
CMD [ "/src/tg-gemini-bot" ]
