FROM registry.fedoraproject.org/fedora:34
ENTRYPOINT [ "/entrypoint.sh" ]

RUN dnf install --setopt=tsflags=nodocs -y \
        postgresql \
        golang-github-cloudflare-cfssl

COPY . .
