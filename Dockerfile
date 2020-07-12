FROM hashicorp/terraform:0.12.28 AS terraform

FROM golang:1.14.4-alpine3.12
RUN apk --no-cache add make git bash
COPY --from=terraform /bin/terraform /usr/local/bin/
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install
