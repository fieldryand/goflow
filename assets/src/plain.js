function updateTaskStateCircles(jobRuns) {
  for (i in jobRuns) {
    taskState = jobRuns[i].jobState.taskState;
    for (taskName in taskState) {
      taskRunStates = jobRuns.map(gettingJobRunTaskState(taskName)).join("");
      document.getElementById(taskName).innerHTML = taskRunStates;
    }
  }
}

function updateGraphViz(jobRuns) {
  if (jobRuns.length) {
    lastJobRun = jobRuns.reverse()[0]
    taskState = lastJobRun.jobState.taskState;
    for (taskName in taskState) {
      if (document.getElementsByClassName("output")) {
        taskRunColor = getJobRunTaskColor(lastJobRun, taskName);
        rect = document.getElementById("node-" + taskName).querySelector("rect");
        rect.setAttribute("style", "stroke-width: 2; stroke: " + taskRunColor);
      }
    }
  }
}

function updateLastRunTs(jobRuns) {
  if (jobRuns.reverse()[0]) {
    lastJobRunTs = jobRuns.reverse()[0].startedAt;
    lastJobRunTsHTML = document.getElementById("last-job-run-ts-wrapper").innerHTML;
    newHTML = lastJobRunTsHTML.replace(/.*/, `Last run: ${lastJobRunTs}`);
    document.getElementById("last-job-run-ts-wrapper").innerHTML = newHTML;
  }
}

function pollingJobState(jobName) {
  function pollJobState() {
    fetch(`/jobs/${jobName}/jobRuns`)
      .then(response => response.json())
      .then(data => {
        jobRunStates = data.jobRuns.map(getJobRunState).join("");
        document.getElementById(jobName).innerHTML = jobRunStates;
      })
    setTimeout(pollJobState, 2000);
  }
  return pollJobState
}

function pollingJobActive(jobName) {
  function pollJobActive() {
    fetch(`/jobs/${jobName}/isActive`)
      .then(response => response.json())
      .then(data => {
        if (data) {
          document.getElementById("schedule-badge-" + jobName).setAttribute("class", "schedule-badge-active");
        } else {
          document.getElementById("schedule-badge-" + jobName).setAttribute("class", "schedule-badge-inactive");
        }
      })
    setTimeout(pollJobActive, 2000);
  }
  return pollJobActive
}

function pollingTaskState(jobName) {
  function pollTaskState() {
    fetch(`/jobs/${jobName}/jobRuns`)
      .then(response => response.json())
      .then(data => {
        updateTaskStateCircles(data.jobRuns);
        updateGraphViz(data.jobRuns);
        updateLastRunTs(data.jobRuns);
      })
    setTimeout(pollTaskState, 2000);
  }
  return pollTaskState
}

function stateColor(taskState) {
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

  return color
}

function stateCircle(taskState) {
  color = stateColor(taskState);
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

function getJobRunTaskColor(jobRun, task) {
  taskState = jobRun.jobState.taskState[task];
  return stateColor(taskState)
}

function getJobRunState(jobRun) {
  return stateCircle(jobRun.jobState.state)
}

function submit(jobName) {
  var xhttp = new XMLHttpRequest();
  xhttp.open("POST", `/jobs/${jobName}/submit`, true);
  xhttp.send();
}

function toggleActive(jobName) {
  var xhttp = new XMLHttpRequest();
  xhttp.open("POST", `/jobs/${jobName}/toggleActive`, true);
  xhttp.send();
}
