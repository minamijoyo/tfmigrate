ARG TERRAFORM_VERSION=latest
FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform

FROM alpine:3.23 AS opentofu
ARG OPENTOFU_VERSION=latest
ADD https://get.opentofu.org/install-opentofu.sh /install-opentofu.sh
RUN chmod +x /install-opentofu.sh
RUN apk add gpg gpg-agent
RUN ./install-opentofu.sh --install-method standalone --opentofu-version $OPENTOFU_VERSION --install-path /usr/local/bin --symlink-path -

FROM golang:1.26-alpine3.23
RUN apk --no-cache add make git bash
COPY --from=terraform /bin/terraform /usr/local/bin/
COPY --from=opentofu /usr/local/bin/tofu /usr/local/bin/
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install

ENTRYPOINT ["./entrypoint.sh"]
