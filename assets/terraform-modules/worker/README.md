# Worker Terraform module

This Terraform module aims to be a reusable module for generating worker node Ignition
configurations.

It builds on top of the [node](../node) module and adds some worker-specific settings on top of it,
like:
- kubeconfig file for the kubelet
- iscsid service and bind-mounts for the kubelet container
- default worker node labels

Additionally, it exposes various input variables, which allows adding platform-specific changes to the
Ignition configuration.
