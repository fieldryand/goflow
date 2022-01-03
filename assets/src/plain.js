function updateTaskStateCircles(jobRuns) {
  for (i in jobRuns) {
    taskState = jobRuns[i].jobState.taskState.internal;
    for (taskName in taskState) {
      taskRunStates = jobRuns.map(gettingJobRunTaskState(taskName)).join("");
      document.getElementById(taskName).innerHTML = taskRunStates;
    }
  }
}

function updateGraphViz(jobRuns) {
  if (jobRuns.length) {
    lastJobRun = jobRuns.reverse()[0]
    taskState = lastJobRun.jobState.taskState.internal;
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

function updateJobActive(jobName) {
  fetch(`/jobs/${jobName}/isActive`)
    .then(response => response.json())
    .then(data => {
      if (data) {
        document.getElementById("schedule-badge-" + jobName).setAttribute("class", "schedule-badge-active-true");
      } else {
        document.getElementById("schedule-badge-" + jobName).setAttribute("class", "schedule-badge-active-false");
      }
    })
}

async function updateAllJobStateCircles() {
  await fetch(`/jobs`)
    .then(response => response.json())
    .then(data => {
      for (i in data) {
        job = data[i]
        updateJobStateCircles(job)
      }
    })
}

function updateJobStateCircles(jobName) {
  var stream = new EventSource(`/jobs/${jobName}/jobRuns`);
  stream.addEventListener("message", function(e) {
    data = JSON.parse(e.data);
    jobRunStates = data.jobRuns.map(getJobRunState).join("");
    document.getElementById(jobName).innerHTML = jobRunStates;
  });
}

function readTaskStream(jobName) {
  var stream = new EventSource(`/jobs/${jobName}/jobRuns`);
  stream.addEventListener("message", function(e) {
    jobRuns = JSON.parse(e.data).jobRuns;
    updateTaskStateCircles(jobRuns);
    updateGraphViz(jobRuns);
    updateLastRunTs(jobRuns);
  });
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
    taskState = jobRun.jobState.taskState.internal[task];
    return stateCircle(taskState)
  }
  return getJobRunTaskState
}

function getJobRunTaskColor(jobRun, task) {
  taskState = jobRun.jobState.taskState.internal[task];
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

async function toggleActive(jobName) {
  const options = {
    method: 'POST'
  }
  await fetch(`/jobs/${jobName}/toggleActive`, options)
    .then(updateJobActive(jobName))
}
