kubeconfig := $(KUBECONFIG)
## Following kubeconfig path is only valid from CI
ifeq ($(RUN_FROM_CI),"true")
	kubeconfig := "${HOME}/assets/auth/kubeconfig"
endif

.PHONY: run-e2e-tests
run-e2e-tests: kube-hunter

kube-hunter:
	KUBECONFIG=${kubeconfig} ./scripts/kube-hunter.sh
