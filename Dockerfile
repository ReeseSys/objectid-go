ARG MONGO_VER="5.0.3"
ARG MONGO_IMAGE=mongo:${MONGO_VER}

FROM ${MONGO_IMAGE} as mongo

# Example Go download URL:
# https://go.dev/dl/go1.19.6.linux-amd64.tar.gz
ARG GO_TAR="go1.19.6.linux-amd64.tar.gz"

WORKDIR /
ADD https://go.dev/dl/${GO_TAR} .
RUN apt-get update \
    && apt-get install -y gcc \
    && tar -C /usr/local/ -xzf ${GO_TAR}

ENV PATH="${PATH}:/usr/local/go/bin"

# This will be a shared volume containing the root source
WORKDIR /src

