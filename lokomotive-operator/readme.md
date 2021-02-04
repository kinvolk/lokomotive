## Lokomotive operator

**IMPORTANT**: Currently under experimentation.

### How to run operator locally

```bash
# To build the code.
make all

# This will install CRDs to the cluster.
make install

# To run code for development.
make run ENABLE_WEBHOOKS=false
```

### Creating custom resource

There is an example yaml in the `config` folder, you can apply that. Before running this make sure to edit config field as shown in [CR reference](#CR-reference)

```bash
$ kubectl apply -f config/samples/components_v1_lokomotivecomponent.yaml
```

### CR reference

For CR follow this guide, it should be of this form

```yaml
apiVersion: components.kinvolk.io/v1
kind: LokomotiveComponent
metadata:
  name: # component name e.g httpbin
spec:
  # config should have 2 keys 
  # component-name.lokocfg e.g httpbin.lokocfg
  # lokocfg.vars
  config:
    httpbin.lokocfg: # path to cluster config file e.g /Users/knrt10/dev/kinvolk/lokomotive-infra/mycluster/cluster.lokocfg
    lokocfg.vars: # path to lokocfg.vars file e.g /Users/knrt10/dev/kinvolk/lokomotive-infra/mycluster/lokocfg.vars
```
