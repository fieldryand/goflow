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

function stateCircle(stateColor) {
  return `
  <svg height="20" width="20">
    <circle cx="10" cy="10" r="9" stroke="black" fill="${stateColor}"/>
  </svg>`
}

const stateColorMap = {
  "Running": "lime",
  "UpForRetry": "yellow",
  "Successful": "green",
  "Failed": "red",
}

function gettingJobRunTaskState(task) {
  function getJobRunTaskState(jobRun) {
    taskState = jobRun.jobState.taskState[task];
    stateColor = stateColorMap[taskState];
    return stateCircle(stateColor)
  }
  return getJobRunTaskState
}

function getJobRunState(jobRun) {
  stateColor = stateColorMap[jobRun.jobState.state];
  return stateCircle(stateColor)
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
