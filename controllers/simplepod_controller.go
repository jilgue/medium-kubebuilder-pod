/*
Copyright 2021.

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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mediumv1alpha1 "simplepod/api/v1alpha1"
)

// SimplePodReconciler reconciles a SimplePod object
type SimplePodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=medium.callepuzzle.com,resources=simplepods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=medium.callepuzzle.com,resources=simplepods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=medium.callepuzzle.com,resources=simplepods/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimplePod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SimplePodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var instance mediumv1alpha1.SimplePod
	errGet := r.Get(ctx, req.NamespacedName, &instance)
	if errGet != nil {
		log.Error(errGet, "Error getting instance")
		return ctrl.Result{}, client.IgnoreNotFound(errGet)
	}

	pod := NewPod(&instance)

	_, errCreate := ctrl.CreateOrUpdate(ctx, r.Client, pod, func() error {
		return ctrl.SetControllerReference(&instance, pod, r.Scheme)
	})

	if errCreate != nil {
		log.Error(errCreate, "Error creating pod")
		return ctrl.Result{}, nil
	}

	err := r.Status().Update(context.TODO(), &instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func NewPod(pod *mediumv1alpha1.SimplePod) *corev1.Pod {
	labels := map[string]string{
		"app": pod.Name,
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: strings.Split(pod.Spec.Command, " "),
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimplePodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediumv1alpha1.SimplePod{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
