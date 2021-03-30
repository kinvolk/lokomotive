# Calico Images

This script downloads the calico images from docker hub and then uploads them to Quay.

## Prerequisites

- An AMD architecture based machine with docker installed.
- An ARM architecture based machine with docker installed.

## Run

Execute the following steps once on AMD and then on ARM machine.

### Docker login

Run the following command in terminal:

```bash
docker login -u <username> quay.io
```

### Execute script

Run the following command in terminal:

```bash
./calico-images.sh v3.18.1
```

> **NOTE:** If the script fails with error `Run the script on AMD machine as well!` or `Run the script on ARM machine as well!` then just run the script on other architecture.
