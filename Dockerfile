FROM ubuntu:18.04

COPY dns-server /bin/dns-server

EXPOSE 53/udp

EXPOSE 8000

ENTRYPOINT ["/bin/dns-server"]
