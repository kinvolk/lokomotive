---
title: DNS
weight: 10
---

Lokomotive relies on DNS records for cluster bootstrap as well as routine operation. When choosing
one of the [supported DNS providers](#supported-providers), Lokomotive can provision the necessary
DNS records on your behalf.

## Supported providers

### [AWS Route 53](https://aws.amazon.com/route53/)

The AWS Route 53 DNS provider is supported with AWS and Packet clusters.

#### IAM permissions

The AWS Route 53 DNS provider requires the following IAM permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "route53:GetChange",
                "route53:GetHostedZone",
                "route53:ChangeResourceRecordSets",
                "route53:ListResourceRecordSets",
                "route53:ListTagsForResource"
            ],
            "Resource": [
                "arn:aws:route53:::change/*",
                "arn:aws:route53:::hostedzone/<HOSTED_ZONE_ID>"
            ]
        },
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": "route53:ListHostedZones",
            "Resource": "*"
        }
    ]
}
```

### [Cloudflare](https://www.cloudflare.com/dns/)

The Cloudflare DNS provider is supported with Packet clusters.

### Manual

The Manual DNS provider is supported with Packet clusters.

This is a special provider. When used, no DNS records are provisioned
automatically. Instead, the user is prompted to configure the necessary DNS
records on their own.

## Records

The records on which Lokomotive relies are mostly DNS **A records**. They are constructed based on
the DNS zone provided by the user and the configured cluster name in the following format:

    <record>.<cluster_name>.<zone>

The exact set of records created varies slightly based on the platform on top of which Lokomotive
is deployed. The following sections provide a per-platform explanation.

## AKS

AKS is a managed Kubernetes platform which takes care of any required DNS configuration on the
user's behalf. Consequently, Lokomotive doesn't manage DNS for AKS deployments.

## AWS

Lokomotive deployments on AWS leverage an AWS
[Network Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/introduction.html)
(NLB). This load balancer serves a dual purpose by default: routing Kubernetes API traffic to the
controller nodes and routing HTTP/S traffic to ingress pods.

Following is a sample record set for an AWS cluster called `my-cluster` with 3 controller nodes
under a zone called `example.com`. For illustration, assume the controller nodes have the following
IP addresses:

- Controller 1: `35.0.0.1` (public) and `10.0.0.1` (private)
- Controller 2: `35.0.0.2` (public) and `10.0.0.2` (private)
- Controller 3: `35.0.0.3` (public) and `10.0.0.3` (private)

Based on the information above, the following DNS records are created:

- `my-cluster.example.com` - an
  [alias record](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-choosing-alias-non-alias.html)
  which automatically resolves to one of the NLB's IP addresses
- `my-cluster-etcd0.example.com` - resolves to `10.0.0.1`
- `my-cluster-etcd1.example.com` - resolves to `10.0.0.2`
- `my-cluster-etcd2.example.com` - resolves to `10.0.0.3`
- `*.my-cluster.example.com` - a CNAME record which resolves to `my-cluster.example.com`

>NOTE: The wildcard CNAME record allows to configure arbitrary subdomains on an ingress controller.
>This enables exposing Kubernetes services to the internet automatically, without having to
>configure an individual DNS record for each service.

## Bare Metal

Lokomotive currently doesn't support DNS provisioning for bare metal clusters. The user has to
create the necessary records *in advance*, before deploying a cluster.

## Packet

Following is a sample record set for a Packet cluster called `my-cluster` with 3 controller nodes
under a zone called `example.com`. For illustration, assume the controller nodes have the following
IP addresses:

- Controller 1: `147.0.0.1` (public) and `10.0.0.1` (private)
- Controller 2: `147.0.0.2` (public) and `10.0.0.2` (private)
- Controller 3: `147.0.0.3` (public) and `10.0.0.3` (private)

Based on the information above, the following DNS records are created:

- `my-cluster.example.com` - resolves to the following IPs in round robin: `147.0.0.1`, `147.0.0.2`, `147.0.0.3`
- `my-cluster-private.example.com` - resolves to the following IPs in round robin: `10.0.0.1`, `10.0.0.2`, `10.0.0.3`
- `my-cluster-etcd0.example.com` - resolves to `10.0.0.1`
- `my-cluster-etcd1.example.com` - resolves to `10.0.0.2`
- `my-cluster-etcd2.example.com` - resolves to `10.0.0.3`
