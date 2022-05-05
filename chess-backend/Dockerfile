FROM node:16-alpine3.15 as build-node
#RUN apk --no-cache --virtual build-dependencies add \
        #python \
        #make \
       # g++

WORKDIR /workdir
#RUN ls ../
COPY chess-board/public ./public
COPY chess-board/src ./src
COPY chess-board/package-lock.json .
COPY chess-board/package.json .
RUN npm install
RUN npm run build

FROM golang:1.18-alpine3.15 as build-go

WORKDIR /app
ENV GOPATH ""
RUN go env -w GOPROXY=direct
RUN apk add git

ADD chess-backend/go.mod chess-backend/go.sum ./
COPY chess-backend/main.go ./
COPY chess-backend/moveset ./moveset
RUN go mod download
ADD . .
RUN go build -o /main

FROM alpine:3.15
COPY --from=build-go /main /main
COPY --from=build-node /workdir/build ./static/
ENTRYPOINT [ "/main" ]