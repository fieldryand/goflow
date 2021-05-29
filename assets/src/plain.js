function pollingJobState(jobName) {
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

function pollingTaskState(jobName) {
  function pollTaskState() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
        var jobRuns = JSON.parse(this.response).jobRuns;
        for (i in jobRuns) {
          taskState = jobRuns[i].jobState.taskState;
          for (taskName in taskState) {
            var taskRunStates = jobRuns.map(gettingJobRunTaskState(taskName)).join("");
            document.getElementById(taskName).innerHTML = taskRunStates;
          }
        }
      }
    };
    xhttp.open("GET", `/jobs/${jobName}/jobRuns`, true);
    xhttp.send();
    setTimeout(pollTaskState, 2000);
  }
  return pollTaskState
}

function stateCircle(taskState) {
  switch (taskState) {
    case "Running":
      color = "#dffbe3";
      break;
    case "UpForRetry":
      color = "#ffc620";
      break;
    case "Successful":
      color = "#39c84e";
      break;
    case "Skipped":
      color = "#abbefb";
      break;
    case "Failed":
      color = "#ff4020";
      break;
    case "None":
      color = "white";
      break;
  }

  return `
  <div class="status-indicator" style="background-color:${color};"></div>
  `
}

function gettingJobRunTaskState(task) {
  function getJobRunTaskState(jobRun) {
    taskState = jobRun.jobState.taskState[task];
    return stateCircle(taskState)
  }
  return getJobRunTaskState
}

function getJobRunState(jobRun) {
  return stateCircle(jobRun.jobState.state)
}

function getDag(jobName) {
  var xhttp = new XMLHttpRequest();
  xhttp.open("GET", `/jobs/${jobName}/dag`, false);
  xhttp.send();
  return xhttp.responseText;
}

function submit(jobName) {
  var xhttp = new XMLHttpRequest();
  xhttp.open("POST", `/jobs/${jobName}/submit`, true);
  xhttp.send();
}
