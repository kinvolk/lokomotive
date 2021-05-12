# Lokomotive Kubelet Image

This image is built on top of the Typhoon Kubelet image. This image includes a wrapper script that runs `isciadm` in the host namespace. This ensures that all the `isciadm` libraries are loaded from the host.

We also use `sed` and `bash` of the image to run the `ca-syncer` init container.

## Updating Image Tag

In this directory run the following command to update the Kubernetes version:

```bash
export NEW_VERSION=
```

```bash
export CURRENT_VERSION=$(grep ^FROM Dockerfile | cut -d":" -f2)
sed -i "s|$CURRENT_VERSION|$NEW_VERSION|g" Dockerfile
```

## Building Image

```bash
export IMAGE_URL=quay.io/kinvolk/kubelet
```

### x86

```bash
export ARCH=amd64

docker build -t $IMAGE_URL:$NEW_VERSION-$ARCH . && docker push $IMAGE_URL:$NEW_VERSION-$ARCH
```

### ARM

```bash
export ARCH=arm64

docker build -t $IMAGE_URL:$NEW_VERSION-$ARCH . && docker push $IMAGE_URL:$NEW_VERSION-$ARCH
```

### Combined image tag

Now make sure you have `"experimental": "enabled"` in your
`~/.docker/config.json` (surrounded by `{` and `}` if the file is otherwise
empty).

When all images are built on the respective architectures and pushed they can
be combined through a manifest to build a multiarch image:

```bash
docker manifest create $IMAGE_URL:$NEW_VERSION \
    --amend $IMAGE_URL:$NEW_VERSION-amd64 \
    --amend $IMAGE_URL:$NEW_VERSION-arm64

docker manifest annotate $IMAGE_URL:$NEW_VERSION \
    $IMAGE_URL:$NEW_VERSION-amd64 --arch=amd64 --os=linux

docker manifest annotate $IMAGE_URL:$NEW_VERSION \
    $IMAGE_URL:$NEW_VERSION-arm64 --arch=arm64 --os=linux

docker manifest push $IMAGE_URL:$NEW_VERSION
```

> **NOTE**: Above commands can be run from any machine.
