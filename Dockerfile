FROM golang:latest
ARG GITHUB_SHA

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app .

EXPOSE 8080
ENV GITHUB_SHA=${GITHUB_SHA}

RUN chmod +x start.sh
CMD ["./start.sh"]