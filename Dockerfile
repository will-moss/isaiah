FROM scratch

COPY isaiah /

ENV DOCKER_RUNNING=true

ENTRYPOINT ["./isaiah"]
