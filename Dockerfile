FROM debian:stretch-slim

WORKDIR /

COPY bin/yoda-scheduler /usr/local/bin

CMD ["syr-scheduler"]