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
	alertproviderv1 "alertojon.io/pagerduty-operator/api/v1"
	"context"
	"errors"
	"fmt"
	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=alertprovider.alertojon.io,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=alertprovider.alertojon.io,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=alertprovider.alertojon.io,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// get the user object from context
	user := &alertproviderv1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// fmt print user object
	fmt.Printf("The user spec ==== %+v\n", user.Spec)

	pagerdutyAuthToken := os.Getenv("PAGERDUTY_API_KEY")

	if pagerdutyAuthToken == "" {
		return ctrl.Result{}, errors.New("PAGERDUTY_API_KEY is not set")
	}

	pagerDutyClient := pagerduty.NewClient(pagerdutyAuthToken)

	userResponse, pagerdutyErr := pagerDutyClient.CreateUser(pagerduty.User{
		Name:  user.Spec.FirstName + " " + user.Spec.LastName,
		Email: user.Spec.Email,
		Role:  "user",
	})
	if pagerdutyErr != nil {
		var aerr pagerduty.APIError

		if errors.As(err, &aerr) {
			if aerr.RateLimited() {
				fmt.Println("rate limited")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			fmt.Println("unknown status code:", aerr.StatusCode)

			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	fmt.Printf("Pagerduty response ==== %+v\n", userResponse)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertproviderv1.User{}).
		Complete(r)
}
