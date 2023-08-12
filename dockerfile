FROM scratch

WORKDIR /
COPY ./bin/app /app
ENTRYPOINT [ "/app" ]