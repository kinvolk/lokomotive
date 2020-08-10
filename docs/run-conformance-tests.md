---
title: Conformance tests
weight: 10
---

This document enumerates the steps required to run conformance tests for various platforms supported by Lokomotive.

**Note**: There is only one caveat to consider when running tests for AWS. For other platforms you can run conformance tests without making special arrangements.

## AWS

For AWS you need to make sure that node ports are allowed in the security group. To do so make sure you set the `expose_nodeports` cluster property to `true` in the AWS config. Read more about this flag in the [AWS reference docs](configuration-reference/platforms/aws.md).

To install the cluster on AWS follow the [AWS quick start guide](quickstarts/aws.md).

## Running conformance tests

Follow the canonical document [here](https://github.com/cncf/k8s-conformance/blob/master/instructions.md) which instructs on installing sonobuoy and running tests.
