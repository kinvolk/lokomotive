FROM ubuntu:18.04
ENTRYPOINT [ "/entrypoint.sh" ]

RUN apt-get update && apt-get install -y \
  ca-certificates \
  postgresql-client \
  golang-cfssl

COPY . .
