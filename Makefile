kubeconfig := $(KUBECONFIG)
## Following kubeconfig path is only valid from CI
ifeq ($(RUN_FROM_CI),"true")
	kubeconfig := "${HOME}/assets/auth/kubeconfig"
endif

.PHONY: run-e2e-tests
run-e2e-tests: kube-hunter
	KUBECONFIG=${kubeconfig} ./scripts/check-version-skew.sh

kube-hunter:
	KUBECONFIG=${kubeconfig} ./scripts/kube-hunter.sh

.PHONY: update-terraform-render-bootkube
update-terraform-render-bootkube:
	./scripts/update-terraform-render-bootkube.sh $(VERSION)
