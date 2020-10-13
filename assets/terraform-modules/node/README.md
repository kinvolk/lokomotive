# Node Terraform module

This Terraform module aims to be a base for [worker](../worker) and [controller](../controller) Terraform
modules by providing common parts of the Ignition configuration.

It is not instantiated by those modules, but linked using symlinks, to avoid instantiating deeply nested
Terraform modules.

Additionally, it exposes various input variables, which allow to add worker and controller specific changes
to the configuration.

The main use case is to provide extra snippets using the `clc_snippets` variable.
