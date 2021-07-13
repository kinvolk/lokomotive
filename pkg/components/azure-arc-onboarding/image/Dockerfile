FROM mcr.microsoft.com/azure-cli:2.25.0

ENV BASE_URL="https://get.helm.sh"

ENV TAR_FILE="helm-v3.6.1-linux-amd64.tar.gz"

RUN wget ${BASE_URL}/${TAR_FILE} && \
    tar -xvf ${TAR_FILE} && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    rm -rf linux-amd64 && \
    rm -rf ${TAR_FILE}

RUN az extension add --name connectedk8s

WORKDIR /usr/local/bin

COPY azure-arc.sh /usr/local/bin/azure-arc.sh

ENTRYPOINT ["azure-arc.sh"]
