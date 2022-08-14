/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/cmwylie19/swo/api/v1alpha1"
)

// SecretWatcherReconciler reconciles a SecretWatcher object
type SecretWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.caseywylie.io,resources=secretwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.caseywylie.io,resources=secretwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.caseywylie.io,resources=secretwatchers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *SecretWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// SecretWatcher operator instance
	secretWatcher := &apiv1alpha1.SecretWatcher{}
	err := r.Get(ctx, req.NamespacedName, secretWatcher)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("SecretWatcher resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}

	port_int, port_err := strconv.Atoi(secretWatcher.Spec.Port)
	if port_err != nil {
		log.Error(port_err, "Failed to convert port to int")
	}

	// Check if ServiceAccount already exists, if not create a new one
	serviceAccount := &corev1.ServiceAccount{}
	err = r.Get(ctx, types.NamespacedName{Name: secretWatcher.Name, Namespace: secretWatcher.Namespace}, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ServiceAccount
		serviceAccount = r.newServiceAccount(secretWatcher)
		log.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.Create(ctx, serviceAccount)
		if err != nil {
			log.Error(err, "Failed to create new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
			return ctrl.Result{}, err
		}
		// ServiceAccount created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount")
		return ctrl.Result{}, err
	}

	// Check if ClusterRole already exists, if not create a new one
	clusterRole := &rbacv1.ClusterRole{}
	err = r.Get(ctx, types.NamespacedName{Name: secretWatcher.Name, Namespace: secretWatcher.Namespace}, clusterRole)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ClusterRole
		clusterRole = r.newClusterRole(secretWatcher)
		log.Info("Creating a new ClusterRole", "ClusterRole.Name", clusterRole.Name)
		err = r.Create(ctx, clusterRole)
		if err != nil {
			log.Error(err, "Failed to create new ClusterRole", "ClusterRole.Name", clusterRole.Name)
			return ctrl.Result{}, err
		}
		// ClusterRole created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRole")
		return ctrl.Result{}, err
	}

	// Check if ClusterRoleBinding already exists, if not create a new one
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: secretWatcher.Name, Namespace: secretWatcher.Namespace}, clusterRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ClusterRoleBinding
		clusterRoleBinding = r.newClusterRoleBinding(secretWatcher)
		log.Info("Creating a new ClusterRoleBinding", "ClusterRoleBinding.Name", clusterRoleBinding.Name)
		err = r.Create(ctx, clusterRoleBinding)
		if err != nil {
			log.Error(err, "Failed to create new ClusterRoleBinding", "ClusterRoleBinding.Name", clusterRoleBinding.Name)
			return ctrl.Result{}, err
		}
		// ClusterRoleBinding created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRoleBinding")
		return ctrl.Result{}, err
	}

	// Check if Service already exists, if not create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: secretWatcher.Name, Namespace: secretWatcher.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		service = r.newService(secretWatcher, port_int)
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: secretWatcher.Name, Namespace: secretWatcher.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeployment(secretWatcher, port_int)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	replicas := secretWatcher.Spec.Replicas
	if *found.Spec.Replicas != replicas {
		found.Spec.Replicas = &replicas
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Ensure deployment has same label as the spec
	label := secretWatcher.Spec.Label
	if found.Spec.Template.Spec.Containers[0].Command[5] != label {
		found.Spec.Template.Spec.Containers[0].Command[5] = label
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Ensure deployment has same port as the spec
	if found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != int32(port_int) {
		found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = int32(port_int)
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.SecretWatcher{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// Return labels for secretwatcher manifests
func labelsForSecretWatcher(name string) map[string]string {
	return map[string]string{"app": "secret-watcher", "secretwatcher_cr": name}
}

// New Service for SecretWatcher
func (r *SecretWatcherReconciler) newService(sw *apiv1alpha1.SecretWatcher, int_port int) *corev1.Service {
	labels := labelsForSecretWatcher(sw.Name)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sw.Name,
			Namespace: sw.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(int_port),
					TargetPort: intstr.IntOrString{
						IntVal: int32(int_port),
					},
				},
			},
			Selector: labels,
		},
	}
	// Set SecretWatcher instance as the owner and controller
	ctrl.SetControllerReference(sw, service, r.Scheme)
	return service
}

// New SecretWatcherServiceAccount
func (r *SecretWatcherReconciler) newServiceAccount(sw *apiv1alpha1.SecretWatcher) *corev1.ServiceAccount {
	labels := labelsForSecretWatcher(sw.Name)
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sw.Name,
			Namespace: sw.Namespace,
			Labels:    labels,
		},
	}
}

// new ClusterRoleBinding for SecretWatcher
func (r *SecretWatcherReconciler) newClusterRoleBinding(sw *apiv1alpha1.SecretWatcher) *rbacv1.ClusterRoleBinding {
	labels := labelsForSecretWatcher(sw.Name)
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sw.Name,
			Namespace: sw.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sw.Name,
				Namespace: sw.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     sw.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

// New ClusterRole
func (r *SecretWatcherReconciler) newClusterRole(sw *apiv1alpha1.SecretWatcher) *rbacv1.ClusterRole {
	labels := labelsForSecretWatcher(sw.Name)
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sw.Name,
			Namespace: sw.Namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

// New secretwatcher deployment
func (r *SecretWatcherReconciler) newDeployment(m *apiv1alpha1.SecretWatcher, port_int int) *appsv1.Deployment {
	ls := labelsForSecretWatcher(m.Name)
	replicas := m.Spec.Replicas
	port := m.Spec.Port
	label := m.Spec.Label

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{

					Containers: []corev1.Container{{
						Image:   "cmwylie19/secret-watcher:latest",
						Name:    "secret-watcher",
						Command: []string{"./secret-watcher", "server", "-p", port, "-l", label},

						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/health",
									Port: intstr.FromInt(port_int),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/health",
									Port: intstr.FromInt(port_int),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},
						Resources: m.Spec.Resources,
					}},
				},
			},
		},
	}

	// Set SecretWatcher instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
