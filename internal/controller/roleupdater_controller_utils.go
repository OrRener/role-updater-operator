package controller

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	roleupdaterv1 "github.com/OrRener/role-updater-operator/api/v1"
	configv1 "github.com/openshift/api/config/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if err := r.Get(ctx, client.ObjectKey{Name: configMapRef.Name, Namespace: configMapRef.Namespace}, configMap); err != nil {
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

func (r *RoleUpdaterReconciler) createAndUpdateStatus(ctx context.Context, instance *roleupdaterv1.RoleUpdater,
	status, message, lastChecked, clusterVersion string, present bool) error {

	instance.Status = roleupdaterv1.RoleUpdaterStatus{
		Status:           status,
		Message:          message,
		LastCheckTime:    lastChecked,
		ClusterVersion:   clusterVersion,
		ConfigMapPresent: present,
	}

	if err := r.Status().Update(ctx, instance); err != nil {
		return err
	}
	return nil
}

func (r *RoleUpdaterReconciler) createJob(ctx context.Context, instance *roleupdaterv1.RoleUpdater, script string) error {

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-job",
			Namespace: "ocp-role-updater-controller",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: func(i int32) *int32 { return &i }(10),
			BackoffLimit:            func(i int32) *int32 { return &i }(3),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "role-updater-operator-controller-manager",
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    instance.Name + "-container",
							Image:   "quay.io/openshift/origin-cli:latest",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{fmt.Sprintf("%s | oc apply -f -", script)},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, job); err != nil {
		return err
	}

	return nil
}

func (r *RoleUpdaterReconciler) getJobStaus(ctx context.Context, instance *roleupdaterv1.RoleUpdater) (string, error) {
	job := &batchv1.Job{}
	if err := r.Get(ctx, client.ObjectKey{Name: instance.Name + "-job", Namespace: "ocp-role-updater-controller"}, job); err != nil {
		if apierrors.IsNotFound(err) {
			return "Missing", nil
		}
		return "", err
	}

	if job.Status.Succeeded > 0 {
		return "Succeeded", nil
	} else if job.Status.Failed > 0 && job.Status.Failed >= *job.Spec.BackoffLimit {
		return "Failed", nil
	} else {
		return "Running", nil
	}
}
