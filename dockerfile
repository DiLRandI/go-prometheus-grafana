FROM scratch

WORKDIR /
COPY ./bin/app /app
COPY data.json /data.json
ENTRYPOINT [ "/app" ]