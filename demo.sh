curl localhost:8090/job/example/submit &&
curl localhost:8090/status &&
sleep 1 &&
curl localhost:8090/status &&
sleep 2 &&
curl localhost:8090/status
