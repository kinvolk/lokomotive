#!/bin/bash

git_version=$1
if [ -z $git_version ]; then
    echo "Please provide the terraform-render-bootkube git version, you can find it at https://github.com/kinvolk/terraform-render-bootkube/"
    echo "make VERSION=\"06bc0786082cc5203ff8867246b1bcd6c1612d83\" update-terraform-render-bootkube"
    exit 1
fi

set -euo pipefail

repo="\"github.com/kinvolk/terraform-render-bootkube?ref=${git_version}\""

for f in $(find ./ -name bootkube.tf)
do
  sed -i "/source =/c\  source = ${repo}" $f
done
