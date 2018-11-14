# Manifests

This tool uses Helm charts to add new components. To add a new component `foo`:

1. Create a new directory under `manifests` of the same name.
2. Create a file called `Chart.yaml` under `manifests/foo`. You can read more about the file here: https://docs.helm.sh/developing_charts#the-chart-yaml-file
3. Create a new directory called `templates` under `manifests/foo`.
4. Place all your Kubernetes manifests under `manifests/foo/templates/`. Ensure that all the manifests are in the same directory level and not nested in sub directories. This is a limitation of Helm not being able to work with charts with manifests in sub directories at the moment. If the component needs to be installed in a specific namespace or needs CRDs to function, ensure that they are in different files as the component installer checks for manifests and applies them first in order before moving onto other parts.
5. Run `make bindata-components build` from the project root.
6. Commit the new changes.
7. Run `./lokoctl components install ...`
