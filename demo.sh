curl localhost:8090/jobs/example/dag &&
echo &&
curl localhost:8090/jobs/example/submit &&
echo &&
curl localhost:8090/jobs/example/state &&
echo &&
sleep 1 &&
curl localhost:8090/jobs/example/state &&
echo &&
sleep 2 &&
curl localhost:8090/jobs/example/state
