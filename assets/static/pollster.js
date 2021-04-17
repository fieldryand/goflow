function pollster(jobName) {
  function pollJobState() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
        var parsed = JSON.parse(this.response)
        var jobRunStates = parsed.jobRuns.map(getJobRunState).join("");
        document.getElementById(jobName).innerHTML = jobRunStates;
      }
    };
    xhttp.open("GET", `/jobs/${jobName}/jobRuns`, true);
    xhttp.send();
    setTimeout(pollJobState, 2000);
  }
  return pollJobState
}

function getJobRunState(jobRun) {
  stateColorMap = {
    "Running": "lime",
    "UpForRetry": "yellow",
    "Successful": "green",
    "Failed": "red",
  };
  stateColor = stateColorMap[jobRun.jobState.state];
  stateCircle = `
    <svg height="20" width="20">
      <circle cx="10" cy="10" r="10" fill="${stateColor}"/>
    </svg>`;
  return stateCircle
}
