FROM ubuntu:20.04

COPY ./cortex /cortex

EXPOSE 9009

ENTRYPOINT [ "/cortex" ]
