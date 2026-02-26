# ---- BUILD IMAGE FOR LAMBDA -----
FROM golang:1.24 AS build

WORKDIR /lambda
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN env GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main main.go

# ---- RUNTIME IMAGE FOR LAMBDA ---
FROM public.ecr.aws/lambda/provided:al2023

COPY --from=build /lambda/main ./main

ENTRYPOINT [ "/usr/local/bin/aws-lambda-rie", "./main" ]
