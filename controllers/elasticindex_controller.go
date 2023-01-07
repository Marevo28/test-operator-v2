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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

	elasticIndex := &appsv1.ElasticIndex{}

	if err := r.Client.Get(ctx, req.NamespacedName, elasticIndex); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ElasticIndex resource not found. Ignoring since object must be deleted")

			log.Info("response of delete: ", req.Name)
			response, err := DeleteIndex(ctx, req.Name)
			if err != nil {
				log.Error(err, "error while deleting elasticIndex", "indexName", elasticIndex.Spec.Name)
			} else {
				log.Info("response of delete: ", response)
			}
			log.Info("ElasticIndex ", elasticIndex.Spec.Name, " has been deleted")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Error(err, "Failed to get ElasticIndex")
	}
	if deleteRequest, err := deleteExternalResources(ctx, elasticIndex, r); err != nil {
		return ctrl.Result{}, err
	} else if !deleteRequest {
		log.Info("create/update ElasticIndex", "indexName ", elasticIndex.Spec.Name)
		code, err := CreateUpdateIndex(elasticIndex)
		if err != nil {
			log.Error(err, "error while updating elasticIndex", "indexName", elasticIndex.Spec.Name)
		} else if code == 200 {
			log.Info("Индекс успешно обновлён")
		} else {
			log.Info("Не удалось обновить индекс", elasticIndex.Name, "Код запроса:", code)
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticIndexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1.ElasticIndex{}).
		Complete(r)
}

func CreateUpdateIndex(ha *appsv1.ElasticIndex) (int, error) {
	name := ha.Spec.Name
	settings := ha.Spec.Settings
	client := &http.Client{}
	var datatest = strings.NewReader(settings)
	req, err := http.NewRequest("PUT", "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/"+name, datatest)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	code := resp.StatusCode
	return code, err
}

func DeleteIndex(ctx context.Context, indexName string) (int, error) {
	client := &http.Client{}
	url := "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/" + indexName
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	code := resp.StatusCode
	return code, err
}

func deleteExternalResources(ctx context.Context, elasticIndex *appsv1.ElasticIndex, r *ElasticIndexReconciler) (bool, error) {
	deleteRequest := false
	finalizerName := "elastic.test.com/finalizer"

	if elasticIndex.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(elasticIndex, finalizerName) {
			log.Info("register a finalizer")
			elasticIndex.ObjectMeta.Finalizers = append(elasticIndex.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, elasticIndex); err != nil {
				return deleteRequest, err
			}
		}
	} else {
		log.Info("elasticindex is being deleted")
		deleteRequest = true
		if controllerutil.ContainsFinalizer(elasticIndex, finalizerName) {
			if _, err := DeleteIndex(ctx, elasticIndex.Spec.Name); err != nil {
				log.Error(err, "error while deleting elasticIndex", "indexName", elasticIndex.Spec.Name)
			}
			// remove finalizer from the list and update it.
			controllerutil.RemoveFinalizer(elasticIndex, finalizerName)
			if err := r.Update(ctx, elasticIndex); err != nil {
				return deleteRequest, err
			}
		}
	}
	log.Info("Status deleteRequest: ", deleteRequest)
	return deleteRequest, nil
}
