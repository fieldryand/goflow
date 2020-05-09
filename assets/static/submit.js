function submit(jobName) {
  var xhttp = new XMLHttpRequest();
  xhttp.open("GET", `/jobs/${jobName}/submit`, true);
  xhttp.send();
}
