ARG TERRAFORM_VERSION=1.12.2
FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform
FROM alpine/terragrunt:latest AS terragrunt


FROM golang:1.24-alpine3.21
RUN apk --no-cache add make git bash
COPY --from=terraform /bin/terraform /usr/local/bin/
COPY --from=terragrunt /usr/local/bin/terragrunt /usr/local/bin/
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install


ENTRYPOINT ["./entrypoint.sh"]
