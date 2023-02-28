#!/bin/sh

IMAGE=mongo-go-test
NET=mongo-test-net
TAG="5.0.3"
SERVER=mongo-test

echo "Building test image..."
docker build -t ${IMAGE} .

echo "Setting up environment..."
docker network create ${NET}
docker run --rm --network ${NET} --name ${SERVER} -d ${IMAGE}

echo "Running tests..."
docker run -it --network ${NET} --rm --name ${SERVER}-runner -v $(pwd):/src -e MONGO_ADDR="mongodb://${SERVER}:27017" ${IMAGE} go test ${1}

echo "Cleaning up..."
docker stop ${SERVER}
docker network remove ${NET}

