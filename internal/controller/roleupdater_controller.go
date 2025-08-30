package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	roleupdaterv1 "github.com/OrRener/role-updater-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type RoleUpdaterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=role.updater.compute.io,resources=roleupdaters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=role.updater.compute.io,resources=roleupdaters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=role.updater.compute.io,resources=roleupdaters/finalizers,verbs=update
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func (r *RoleUpdaterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	instance, err := r.fetchInstance(ctx, req)
	if err != nil {
		log.Error(err, "failed to fetch RoleUpdater instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Successfully fetched RoleUpdater instance", "instance:", instance)

	configMap := &corev1.ConfigMap{}
	configMap, err = r.checkIfConfigMapExists(ctx, instance)
	if err != nil {
		r.createAndUpdateStatus(ctx, instance, "Error", err.Error(), time.Now().Format(time.RFC3339), "", false)
		log.Error(err, "failed to fetch configMap")
		return ctrl.Result{}, err
	}
	log.Info("Successfully fetched ConfigMap", "configMap:", configMap)

	clusterVersion, err := r.getClusterVersion(ctx)
	if err != nil {
		r.createAndUpdateStatus(ctx, instance, "Error", err.Error(), time.Now().Format(time.RFC3339), "", true)
		log.Error(err, "failed to fetch cluster version")
		return ctrl.Result{}, err
	}
	log.Info("Successfully fetched ClusterVersion", "clusterVersion:", clusterVersion)

	if !instance.Spec.ForceRun {
		if clusterVersion == instance.Status.ClusterVersion || instance.Status.ClusterVersion == "" {
			r.createAndUpdateStatus(ctx, instance, "Tracking", "this resource watches the cluster version", time.Now().Format(time.RFC3339), "", true)
			log.Info("Watching for changes in the clusterVersion", "instance:", instance)
			return ctrl.Result{RequeueAfter: time.Hour * 1}, nil
		}
	}

	log.Info("this is the script from the configMap", "script:", configMap.Data["script.sh"])
	r.createAndUpdateStatus(ctx, instance, "Executing", "executing the script", time.Now().Format(time.RFC3339), "", true)

	return ctrl.Result{RequeueAfter: time.Hour * 1}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleUpdaterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&roleupdaterv1.RoleUpdater{}).
		Named("roleupdater").
		Complete(r)
}
