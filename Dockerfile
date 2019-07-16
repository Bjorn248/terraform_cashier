FROM alpine:3.7

RUN apk add --update git bash openssh

RUN wget https://github.com/Bjorn248/terraform_cashier/releases/download/0.6/cashier_linux.tar.gz \
    && tar zxf cashier_linux.tar.gz \
    && mkdir /app \
    && mv terraform_cashier_linux_amd64 /app/cashier


FROM alpine:latest

RUN apk --no-cache add ca-certificates

ENV TERRAFORM_PLANFILE="/data/terraform.plan"
ENV AWS_REGION="us-east-1"
WORKDIR /data
COPY --from=0 /app/cashier /app/cashier

ENTRYPOINT [ "/app/cashier" ]