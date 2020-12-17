---
title: Logging
weight: 10
---

## Logging

Lokomotive restricts the logs per Docker container (not Kubernetes pod) to be 300MB divided into
three files (100MB per file). This should allow running around fifty log intensive containers on a
20GB disk. Due to this limitation, any log intensive app will have only last 300MB of logs. Anything
beyond that will be purged by Docker. To preserve logs of such applications for long term access use
a log shipping solution.
