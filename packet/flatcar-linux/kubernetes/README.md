## Install flatcar on packet

#### Create variables' values file

Create a file `terraform.tfvars` with following keys and its corresponding values.

```toml
project_id = ""
ssh_authorized_key = "ssh-rsa ..."
controller_count = "2"
worker_count = "2"
```

- `project_id`(required): This is the id of the project created on your packet account.

- `ssh_keys`(required): List of SSH keys that will be used for password less authentication of `core` user.

- `controller_count`(optional): Number of controller/master nodes to be created. Default is 1.

- `worker_count`(optional): Number of worker nodes to be created. Default is 1.

- `cluster_region`(optional): Provide a datacenter location to install this cluster in. Read [here](https://support.packet.com/kb/articles/data-centers) for the name of datacenters. Default is `ams1`.

#### Deploy servers

Generate the authentication token on the packet website and export it as `PACKET_AUTH_TOKEN`, before running `terraform` commands. Or you can also specify this token when prompted.

```
export PACKET_AUTH_TOKEN=""
terraform init
terraform apply
```

#### Check deployment

Get the IP Address of the servers from the packet console and run following:

```
ssh core@<IP Address>
```

You can ssh and verify that the machine has right hostname.
