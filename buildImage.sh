swagger generate server -f dsb-swagger.yaml -A k8sdsb
docker build -t ocopea/k8s-dsb -f Dockerfile ../
