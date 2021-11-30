ARG TERRAFORM_VERSION=latest
FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform

FROM golang:1.17.3-alpine3.14
RUN apk --no-cache add make git bash
COPY --from=terraform /bin/terraform /usr/local/bin/
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install

ENTRYPOINT ["./entrypoint.sh"]
