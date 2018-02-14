cd ${0%/*}
mkdir ../tmp

CLUSTER_NAME=../tmp/g0
NUM_WORKERS=5
OPENLAMBDA=../../open-lambda # Looks for open-lambda at same directory level as olscheduler repo
ADMIN=$OPENLAMBDA/bin/admin
HANDLERS=$OPENLAMBDA/quickstart/handlers

echo "---> REMOVING CLUSTER: "$CLUSTER_NAME
$ADMIN kill -cluster=$CLUSTER_NAME
rm -r $CLUSTER_NAME
echo 
echo
echo
echo "---> NEW CLUSTER: "$CLUSTER_NAME
$ADMIN new -cluster=$CLUSTER_NAME
$ADMIN workers -n=$NUM_WORKERS -cluster=$CLUSTER_NAME
$ADMIN status -cluster=$CLUSTER_NAME
cp -r $HANDLERS/hello $CLUSTER_NAME/registry/hello
echo 
echo
echo
echo "---> TEST WORKER"
curl -w "/n" -X POST localhost:8080/runLambda/hello -d '{"name": "moon"}'
echo 
echo
echo
echo "---> START BALANCER"
$ADMIN gobetween -cluster=$CLUSTER_NAME
echo 
echo
echo
sleep 1s
echo "---> TEST BALANCER"
curl -w "/n" -X POST localhost:9080/runLambda/hello -d '{"name": "Moon"}'
echo
