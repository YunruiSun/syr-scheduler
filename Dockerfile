FROM debian:stretch-slim

WORKDIR /

COPY bin/syr-scheduler /usr/local/bin

CMD ["syr-scheduler"]