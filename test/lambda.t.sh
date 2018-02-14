cd ${0%/*}
CLUSTER_NAME=../tmp/c0
NUM_WORKERS=5
ADMIN=../bin/admin

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
cp -r ./quickstart/handlers/hello $CLUSTER_NAME/registry/hello
echo 
echo
echo
echo "---> TEST WORKER"
curl -w "\n" -X GET localhost:8080/lambda/hello?cmd=load
curl -w "\n" -X GET localhost:8080/lambda/hello?cmd=scheme
curl -w "\n" -X POST localhost:8080/runLambda/hello -d '{"name": "moon"}'

# curl -w "\n" -X GET localhost:9080
# curl -w "\n" -X GET localhost:9080/scheduler
# curl -w "\n" -X GET localhost:9080/status

curl -w "\n" -X POST localhost:9080/runLambda/hello -d '{"pkgs": ["fmt", "rand"], "name": "Moon"}'
# curl -w "\n" -X POST localhost:9080/runLambda/hello -d '{"pkgs": ["fmt", "rand"]}'
# curl -w "\n" -X POST localhost:9080/runLambda/hello1 -d '{"pkgs": ["strings","errors", "fmt"]}'
# curl -w "\n" -X POST localhost:9080/runLambda/hello2 -d '{"pkgs": ["math", "fmt", "lol"]}'
# curl -w "\n" -X POST localhost:9080/runLambda/hello3 -d '{"pkgs": ["net", "fmt", "lol"]}'