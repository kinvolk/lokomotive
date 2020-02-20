kubeconfig := $(KUBECONFIG)
## Following kubeconfig path is only valid from CI
ifeq ($(RUN_FROM_CI),"true")
	kubeconfig := "${HOME}/assets/auth/kubeconfig"
endif
kubehunter := ./scripts/kube-hunter.sh
ifeq ($(SKIP_KUBE_HUNTER),"true")
	kubehunter := echo
endif

.PHONY: run-e2e-tests
run-e2e-tests: kube-hunter
	KUBECONFIG=${kubeconfig} ./scripts/check-version-skew.sh

kube-hunter:
	KUBECONFIG=${kubeconfig} ${kubehunter}
