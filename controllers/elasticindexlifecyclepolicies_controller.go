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
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	elasticv1 "github.com/Marevo28/test-operator-v2/api/v1"
)

// ElasticIndexLifecyclePoliciesReconciler reconciles a ElasticIndexLifecyclePolicies object
type ElasticIndexLifecyclePoliciesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindexlifecyclepolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindexlifecyclepolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindexlifecyclepolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticIndexLifecyclePolicies object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *ElasticIndexLifecyclePoliciesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("ElasticIndexLifecyclePolicies", req.NamespacedName)

	elasticIndexLifecyclePolicies := &elasticv1.ElasticIndexLifecyclePolicies{}

	if err := r.Client.Get(ctx, req.NamespacedName, elasticIndexLifecyclePolicies); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ElasticIndexLifecyclePolicies resource not found. Ignoring since object must be deleted")
			_, err := DeleteIndexLifecyclePolicies(ctx, req.Name)
			if err != nil {
				log.Error(err, "error while deleting elasticIndexLifecyclePolicies", "indexLifecyclePolicies", elasticIndexLifecyclePolicies.Spec.Name)
			}
			log.Info("ElasticIndexLifecyclePolicies ", elasticIndexLifecyclePolicies.Spec.Name, " has been deleted")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	log.Info("create/update ElasticIndexLifecyclePolicies", "indexLifecyclePolicies ", elasticIndexLifecyclePolicies.Spec.Name)
	code, err := CreateUpdateIndexLifecyclePolicies(elasticIndexLifecyclePolicies)
	if err != nil {
		log.Error(err, "error while updating elasticIndexLifecyclePolicies", "indexLifecyclePolicies", elasticIndexLifecyclePolicies.Spec.Name)
	} else if code == 200 {
		log.Info("Индекс лайф полиси успешно обновлён")
	} else {
		log.Info("Не удалось обновить индекс лайф полиси", elasticIndexLifecyclePolicies.Name, "Код запроса:", code)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticIndexLifecyclePoliciesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1.ElasticIndexLifecyclePolicies{}).
		Complete(r)
}

func CreateUpdateIndexLifecyclePolicies(ha *elasticv1.ElasticIndexLifecyclePolicies) (int, error) {
	name := ha.Spec.Name
	settings := ha.Spec.Settings
	client := &http.Client{}
	var data = strings.NewReader(settings)
	req, err := http.NewRequest("PUT", "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/_ilm/policy/"+name, data)
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

func DeleteIndexLifecyclePolicies(ctx context.Context, indexLifecyclePolicies string) (int, error) {
	client := &http.Client{}
	url := "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/_ilm/policy/" + indexLifecyclePolicies
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
