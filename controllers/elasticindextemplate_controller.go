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

// ElasticIndexTemplateReconciler reconciles a ElasticIndexTemplate object
type ElasticIndexTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindextemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindextemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elastic.test.com,resources=elasticindextemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticIndexTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *ElasticIndexTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("ElasticIndexTemplate", req.NamespacedName)

	elasticIndexTemplate := &elasticv1.ElasticIndexTemplate{}

	if err := r.Client.Get(ctx, req.NamespacedName, elasticIndexTemplate); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ElasticIndexTemplate resource not found. Ignoring since object must be deleted")
			_, err := DeleteIndexTemplate(ctx, req.Name)
			if err != nil {
				log.Error(err, "error while deleting elasticIndexTemplate", "indexTemplateName", elasticIndexTemplate.Spec.Name)
			}
			log.Info("ElasticIndexTemplate ", elasticIndexTemplate.Spec.Name, " has been deleted")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	log.Info("create/update ElasticIndexTemplate", "indexTemplateName ", elasticIndexTemplate.Spec.Name)
	code, err := CreateUpdateIndexTemplate(elasticIndexTemplate)
	if err != nil {
		log.Error(err, "error while updating elasticIndexTemplate", "indexTemplateName", elasticIndexTemplate.Spec.Name)
	} else if code == 200 {
		log.Info("???????????? ???????????????? ?????????????? ????????????????")
	} else {
		log.Info("???? ?????????????? ???????????????? ???????????? ????????????????", elasticIndexTemplate.Name, "?????? ??????????????:", code)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticIndexTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1.ElasticIndexTemplate{}).
		Complete(r)
}

func CreateUpdateIndexTemplate(ha *elasticv1.ElasticIndexTemplate) (int, error) {
	name := ha.Spec.Name
	settings := ha.Spec.Settings
	client := &http.Client{}
	var data = strings.NewReader(settings)
	req, err := http.NewRequest("PUT", "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/_index_template/"+name, data)
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

func DeleteIndexTemplate(ctx context.Context, indexTemplateName string) (int, error) {
	client := &http.Client{}
	url := "http://elasticsearch-master.kube-elasticsearch.svc.cluster.local:9200/_index_template/" + indexTemplateName
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
