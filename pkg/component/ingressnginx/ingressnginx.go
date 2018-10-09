package ingressnginx

import (
	"fmt"

	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

type IngressNginx struct {
}

func (ig IngressNginx) Name() string {
	return "ingress-nginx"
}

func (ig IngressNginx) Install(client *kubernetes.Clientset, namespace string) error {
	contextLogger := log.WithFields(log.Fields{
		"command": fmt.Sprintf("lokoctl component install %s", ig.Name()),
	})

	nsData, err := Asset("manifests/nginx-ingress/0-namespace.yaml")
	if err != nil {
		return err
	}

	nsObj, err := k8sutil.GetKubernetesObjectFromTmpl(nsData, nil)
	if err != nil {
		return err
	}

	newNs := nsObj.(*v1.Namespace)
	ns, err := client.CoreV1().Namespaces().Create(newNs)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			contextLogger.Infof("Namespace %q already exists", newNs.Name)
		} else {
			return err
		}
	} else {
		contextLogger.Infof("Namespace %q created", ns.Name)
	}

	roleData, err := Asset("manifests/nginx-ingress/rbac/cluster-role.yaml")
	if err != nil {
		return err
	}

	roleObj, err := k8sutil.GetKubernetesObjectFromTmpl(roleData, nil)
	if err != nil {
		return err
	}

	newRole := roleObj.(*rbacv1.ClusterRole)
	role, err := client.RbacV1().ClusterRoles().Create(newRole)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			contextLogger.Infof("ClusterRole %q already exists", newRole.Name)
		} else {
			return err
		}
	} else {
		contextLogger.Infof("ClusterRole %q created", role.Name)
	}

	rbData, err := Asset("manifests/nginx-ingress/rbac/cluster-role-binding.yaml")
	if err != nil {
		return err
	}

	rbObj, err := k8sutil.GetKubernetesObjectFromTmpl(rbData, nil)
	if err != nil {
		return err
	}

	newRb := rbObj.(*rbacv1.ClusterRoleBinding)
	rb, err := client.RbacV1().ClusterRoleBindings().Create(newRb)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			contextLogger.Infof("ClusterRoleBinding %q already exists", newRb.Name)
		} else {
			return err
		}
	} else {
		contextLogger.Infof("ClusterRoleBinding %q created", rb.Name)
	}

	depData, err := Asset("manifests/nginx-ingress/deployment.yaml")
	if err != nil {
		return err
	}

	depObj, err := k8sutil.GetKubernetesObjectFromTmpl(depData, nil)
	if err != nil {
		return err
	}

	newDep := depObj.(*appsv1.Deployment)
	dep, err := client.AppsV1().Deployments(newDep.Namespace).Create(newDep)
	if err != nil {
		return err
	}

	contextLogger.Infof("Deployment %q created in namespace %q", dep.Name, dep.Namespace)

	svcData, err := Asset("manifests/nginx-ingress/service.yaml")
	if err != nil {
		return err
	}

	svcObj, err := k8sutil.GetKubernetesObjectFromTmpl(svcData, nil)
	if err != nil {
		return err
	}

	newSvc := svcObj.(*v1.Service)
	svc, err := client.CoreV1().Services(newSvc.Namespace).Create(newSvc)
	if err != nil {
		return err
	}

	contextLogger.Infof("Service %q created in namespace %q", svc.Name, svc.Namespace)

	return nil
}

func New() *IngressNginx {
	return &IngressNginx{}
}
