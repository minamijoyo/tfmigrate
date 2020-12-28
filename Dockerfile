ARG TERRAFORM_VERSION=latest
FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform

FROM golang:1.15.0-alpine3.12
RUN apk --no-cache add make git bash
COPY --from=terraform /bin/terraform /usr/local/bin/
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install

ENTRYPOINT ["./entrypoint.sh"]
