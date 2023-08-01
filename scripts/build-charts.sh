cd ..

helm package charts/rancher-ccg-gmsa-provider

helm package charts/CCG-gMSA-plugin-installer

helm repo index .

cd scripts