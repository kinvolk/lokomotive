# Controller Terraform module

This Terraform module aims to be a reusable module for generating controller node Ignition
configurations.

It builds on top of the [node](../node) module and adds some controller-specific settings on top of it,
like:
- extra `kubelet.service` dependencies etc.
- bootkube script and systemd unit
- etcd scripts and units
- controller labels
- controller taints

Additionally, it exposes various input variables, which allows adding platform-specific changes to the
Ignition configuration.
