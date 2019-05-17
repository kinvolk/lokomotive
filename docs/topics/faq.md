# FAQ

## Terraform

Lokomotive provides a Terraform Module for each supported operating system and platform. Terraform is considered a *format* detail, much like a Linux distro might provide images in the qcow2 or ISO format. It is a mechanism for sharing Lokomotive in a way that works for many users.

Formats rise and evolve. Lokomotive may choose to adapt the format over time (with lots of forewarning). However, the authors' have built several Kubernetes "distros" before and learned from mistakes - Terraform modules are the right format for now.

## Operating Systems

Lokomotive supports Flatcar Linux. This operating system was chosen because it offers:

* Minimalism and focus on clustered operation
* Automated and atomic operating system upgrades
* Declarative and immutable configuration
* Optimization for containerized applications

## Get Help

Ask questions on the IRC #lokomotive-k8s channel on [freenode.net](http://freenode.net/).

## Security Issues

If you find security issues, please see [security disclosures](/topics/security.md#disclosures).

## Maintainers

Lokomotive clusters are Kubernetes clusters the maintainers use in real-world, production clusters.

* Maintainers must personally operate a bare-metal and cloud provider cluster and strive to exercise it in real-world scenarios

We merge features that are along the "blessed path". We minimize options to reduce complexity and matrix size. We remove outdated materials to reduce sprawl. "Skate where the puck is going", but also "wait until the fit is right". No is temporary, yes is forever.
