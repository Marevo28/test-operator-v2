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
	"strings"

	appsv1 "github.com/Marevo28/test-operator-v2/api/v1"
	log "github.com/sirupsen/logrus"

	// "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cri-api/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"encoding/json"
	elasticv1 "github.com/Marevo28/test-operator-v2/api/v1"
	"net/http"
)

// ElasticIndexReconciler reconciles a ElasticIndex object
type ElasticIndexReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticIndex object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *ElasticIndexReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("ElasticIndex", req.NamespacedName)
	log.Println("Processing ElasticIndexReconciler.")
	elasticIndex := &appsv1.ElasticIndex{}
	log.Println(elasticIndex)
	err := r.Client.Get(ctx, req.NamespacedName, elasticIndex)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ElasticIndex resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ElasticIndex")
		return ctrl.Result{}, err
	}
	// Check if the deployment already exists, if not create a new one
	log.Println("Check found")
	// found := &a.Deployment{}
	// log.Println(found)
	// err = r.Client.Get(ctx, types.NamespacedName{Name: elasticIndex.Name, Namespace: elasticIndex.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// dep := r.deployHelloApp(elasticIndex)
	code := CustomCreateIndex(elasticIndex)
	log.Info(code)
	// 	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 	err = r.Client.Create(ctx, dep)
	// 	if err != nil {
	// 		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Deployment created successfully - return and requeue
	// 	return ctrl.Result{Requeue: true}, nil
	// } else if err != nil {
	// 	log.Error(err, "Failed to get Deployment")
	// 	return ctrl.Result{}, err
	// }

	// Check desired amount of deploymets.
	// size := elasticIndex.Spec.Size
	// if *found.Spec.Replicas != size {
	// 	found.Spec.Replicas = &size
	// 	err = r.Client.Update(ctx, found)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Spec updated - return and requeue
	// 	return ctrl.Result{Requeue: true}, nil
	// }
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticIndexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1.ElasticIndex{}).
		Complete(r)
}

func CustomCreateIndex(ha *appsv1.ElasticIndex) int {
	name := ha.Spec.Name
	settings := ha.Spec.Settings

	client := &http.Client{}

	var datatest = strings.NewReader(settings)
	req, err := http.NewRequest("PUT", "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/"+name, datatest)
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	code := resp.StatusCode
	return code
}

type Iot struct {
	Id      int             `json:"id"`
	Name    string          `json:"name"`
	Context json.RawMessage `json:"context"` // RawMessage here! (not a string)
}
