# MetalLB Load Balancer

Install the component by running:

```bash
lokoctl component install metallb
```

Now create `ConfigMap` like following, but change the information as needed of the key `config`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    peers:
    - node-selectors:
      - match-labels:
          kubernetes.io/hostname: suraj-test-worker-0
      peer-address: 10.64.43.10
      peer-asn: 65530
      my-asn: 65000
      hold-time: 5s

    address-pools:
    - name: default
      protocol: bgp

      # elastic IP Addresses block you got from packet
      addresses:
      - 147.75.40.46/32
```

For more information on the configuration of the file above refer the upstream documentation: [https://metallb.universe.tf/configuration/](https://metallb.universe.tf/configuration/).
