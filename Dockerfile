FROM ghcr.io/brandonkowalski/quasimodo:latest

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

COPY . .
RUN GOWORK=off go build -v -gcflags="all=-N -l" -o game-manager app/game_manager.go

CMD ["/bin/bash"]