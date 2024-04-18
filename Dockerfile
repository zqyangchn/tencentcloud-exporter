FROM ubuntu:22.04

LABEL maintainer="zqyangchn@gmail.com" description="qcloud_exporter"

EXPOSE 9123

RUN apt-get -qq update && \
    apt-get -qq install -y --no-install-recommends ca-certificates curl procps iproute2 net-tools iputils-ping telnet && \
    apt -qq autoremove && rm -rf /var/lib/apt/lists/*


ADD configs /etc/qcloud_exporter
ADD qcloud_exporter /opt/qcloud_exporter/qcloud_exporter

WORKDIR /opt/qcloud_exporter/

ENTRYPOINT ["/opt/qcloud_exporter/qcloud_exporter", "--config.file=/etc/qcloud-exporter/qcloud.yml"]
