package controller

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	roleupdaterv1 "github.com/OrRener/role-updater-operator/api/v1"
	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
)

func (r *RoleUpdaterReconciler) fetchInstance(ctx context.Context, req ctrl.Request) (*roleupdaterv1.RoleUpdater, error) {

	instance := &roleupdaterv1.RoleUpdater{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, instance); err != nil {
		return nil, err
	}
	return instance, nil
}

func (r *RoleUpdaterReconciler) checkIfConfigMapExists(ctx context.Context, instance *roleupdaterv1.RoleUpdater) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	configMapRef := instance.Spec.ConfigMapRef
	if configMapRef.Namespace == "" {
		configMapRef.Namespace = instance.Namespace
	}

	if err := r.Get(ctx, client.ObjectKey{Name: configMap.Name, Namespace: configMapRef.Namespace}, configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *RoleUpdaterReconciler) getClusterVersion(ctx context.Context) (string, error) {
	clusterVersion := &configv1.ClusterVersion{}
	if err := r.Get(ctx, client.ObjectKey{Name: "version"}, clusterVersion); err != nil {
		return "", err
	}
	return clusterVersion.Status.Desired.Version, nil
}
