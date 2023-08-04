cd ..

helm package charts/rancher-ccg-dll-installer

helm package charts/rancher-gmsa-account-provider

helm repo index .

cd scripts