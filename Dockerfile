FROM busybox:stable

COPY isaiah /

ENV DOCKER_RUNNING=true

ENTRYPOINT ["./isaiah"]
