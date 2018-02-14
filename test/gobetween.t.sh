cd ${0%/*}
CLUSTER_NAME=../tmp/c0
NUM_WORKERS=7
ADMIN=../bin/admin

echo "---> REMOVING CLUSTER: "$CLUSTER_NAME
./bin/admin kill -cluster=$CLUSTER_NAME
rm -r $CLUSTER_NAME
echo 
echo
echo
echo "---> NEW CLUSTER: "$CLUSTER_NAME
./bin/admin new -cluster=$CLUSTER_NAME
./bin/admin workers -n=$NUM_WORKERS -cluster=$CLUSTER_NAME
./bin/admin status -cluster=$CLUSTER_NAME
cp -r ./quickstart/handlers/hello $CLUSTER_NAME/registry/hello
echo 
echo
echo
echo "---> TEST WORKER"
curl -X POST localhost:8080/runLambda/hello -d '{"name": "moon"}'
echo 
echo
echo
echo "---> START BALANCER"
./bin/admin gobetween -cluster=$CLUSTER_NAME
echo 
echo
echo
sleep 1s
echo "---> TEST BALANCER"
curl -X POST localhost:9080/runLambda/hello -d '{"name": "Moon"}'
echo