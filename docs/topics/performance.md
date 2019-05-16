# Performance

## Provision Time

Provisioning times vary based on the operating system and platform. Sampling the time to create (apply) and destroy clusters with 1 controller and 2 workers shows (roughly) what to expect.

| Platform      | Apply | Destroy |
|---------------|-------|---------|
| AWS           | 5 min | 3 min   |
| Azure         | 10 min | 7 min   |
| Bare-Metal    | 10-15 min | NA  |
| Packet        | 8-15 min | 2 min |

Notes:

* SOA TTL and NXDOMAIN caching can have a large impact on provision time
* Platforms with auto-scaling take more time to provision (AWS, Azure, Google)
* Bare-metal POST times and network bandwidth will affect provision times
