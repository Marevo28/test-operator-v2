## Подготовка окружения
operator-sdk 
operator-sdk init --domain gmcs.ru --repo github.com/Marevo28/test-operator

curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/archive/refs/tags/v3.8.0.tar.gz

curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.8.0/kubebuilder_linux_amd64

chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

kubebuilder version

wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
tar -xvf go1.19.4.linux-amd64.tar.gz   
mv go /usr/local  

nano ~/.bashrc

export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

source ~/.bashrc


## Команды для создания опертора

kubebuilder init --domain test.com --repo github.com/Marevo28/test-operator-v2
kubebuilder create api --group elastic --version v1 --kind ElasticIndex  --resource --controller
kubebuilder create api --group elastic --version v1 --kind ElasticIndexLifecyclePolicies  --resource --controller
kubebuilder create api --group elastic --version v1 --kind ElasticIndexTemplate  --resource --controller
kubebuilder create api --group elastic --version v1 --kind ElasticDataStream  --resource --controller


## Подготовка артефактов
make manifests

make docker-build IMG=marevo28/test-operator:0.0.4

make docker-push IMG=marevo28/test-operator:0.0.4

## Деплой в кластер
kubectl apply -f ..\..\Github\test-operator\config\crd\bases\
kubectl apply -f ..\..\Github\test-operator\config\rbac\role_binding.yaml
kubectl apply -f ..\..\Github\test-operator\config\rbac\role.yaml
kubectl apply -f ..\..\Github\test-operator\config\rbac\service_account.yaml
kubectl apply -f ..\..\Github\test-operator\dist\6-controller.yaml


## Деплой тестового примера
kubectl apply -f ..\..\Github\test-operator\config\samples\elastic_v1_elasticindex.yaml
kubectl apply -f ..\..\Github\test-operator\config\samples\elastic_v1_elasticindexlifecyclepolicies.yaml
kubectl apply -f ..\..\Github\test-operator\config\samples\elastic_v1_elasticindextemplate.yaml
