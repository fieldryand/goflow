function updateTaskStateCircles(jobRuns) {
  for (i in jobRuns) {
    const taskState = jobRuns[i].jobState.taskState.internal;
    for (taskName in taskState) {
      const taskRunStates = jobRuns.map(gettingJobRunTaskState(taskName));
      const old_wrapper = document.getElementById(taskName);
      const new_wrapper = document.createElement("div");
      new_wrapper.setAttribute("class", "status-wrapper");
      new_wrapper.setAttribute("id", taskName);
      for (k in taskRunStates) {
        const color = taskRunStates[k];
        div = document.createElement("div");
        div.setAttribute("class", "status-indicator");
        div.setAttribute("style", `background-color:${color}`);
        new_wrapper.appendChild(div);
      }
      document.getElementById("task-table").replaceChild(new_wrapper, old_wrapper);
    }
  }
}

function updateJobStateCircles() {
  var stream = new EventSource(`/stream`);
  stream.addEventListener("message", function(e) {
    const data = JSON.parse(e.data);
    const jobRunStates = data.jobRuns.map(getJobRunState);
    const old_wrapper = document.getElementById(data.jobName);
    const new_wrapper = document.createElement("div");
    new_wrapper.setAttribute("class", "status-wrapper");
    new_wrapper.setAttribute("id", data.jobName);
    for (k in jobRunStates) {
      const color = jobRunStates[k];
      div = document.createElement("div");
      div.setAttribute("class", "status-indicator");
      div.setAttribute("style", `background-color:${color}`);
      new_wrapper.appendChild(div);
    }
    document.getElementById("job-table").replaceChild(new_wrapper, old_wrapper);
  });
}

function updateGraphViz(jobRuns) {
  if (jobRuns.length) {
    const lastJobRun = jobRuns.reverse()[0]
    const taskState = lastJobRun.jobState.taskState.internal;
    for (taskName in taskState) {
      if (document.getElementsByClassName("output")) {
        const taskRunColor = getJobRunTaskColor(lastJobRun, taskName);
	try {
          const rect = document.getElementById("node-" + taskName).querySelector("rect");
          rect.setAttribute("style", "stroke-width: 2; stroke: " + taskRunColor);
	}
	catch(err) {
          console.log(`${err}. This might be a temporary error when the graph is still loading.`)
	}
      }
    }
  }
}

function updateLastRunTs(jobRuns) {
  if (jobRuns.reverse()[0]) {
    const lastJobRunTs = jobRuns.reverse()[0].startedAt;
    const lastJobRunTsHTML = document.getElementById("last-job-run-ts-wrapper").innerHTML;
    const newHTML = lastJobRunTsHTML.replace(/.*/, `Last run: ${lastJobRunTs}`);
    document.getElementById("last-job-run-ts-wrapper").innerHTML = newHTML;
  }
}

function updateJobActive(jobName) {
  fetch(`/jobs/${jobName}/isActive`)
    .then(response => response.json())
    .then(data => {
      if (data) {
        document
          .getElementById("schedule-badge-" + jobName)
          .setAttribute("class", "schedule-badge-active-true");
      } else {
        document
          .getElementById("schedule-badge-" + jobName)
          .setAttribute("class", "schedule-badge-active-false");
      }
    })
}

function readTaskStream(jobName) {
  var stream = new EventSource(`/stream`);
  stream.addEventListener("message", function(e) {
    const data = JSON.parse(e.data);
    if (jobName == data.jobName) {
      updateTaskStateCircles(data.jobRuns);
      updateGraphViz(data.jobRuns);
      updateLastRunTs(data.jobRuns);
    }
  });
}

function stateColor(taskState) {
  switch (taskState) {
    case "Running":
      var color = "#dffbe3";
      break;
    case "UpForRetry":
      var color = "#ffc620";
      break;
    case "Successful":
      var color = "#39c84e";
      break;
    case "Skipped":
      var color = "#abbefb";
      break;
    case "Failed":
      var color = "#ff4020";
      break;
    case "None":
      var color = "white";
      break;
  }

  return color
}

function gettingJobRunTaskState(task) {
  function getJobRunTaskState(jobRun) {
    const taskState = jobRun.jobState.taskState.internal[task];
    return stateColor(taskState)
  }
  return getJobRunTaskState
}

function getJobRunTaskColor(jobRun, task) {
  const taskState = jobRun.jobState.taskState.internal[task];
  return stateColor(taskState)
}

function getJobRunState(jobRun) {
  return stateColor(jobRun.jobState.state)
}

function submit(jobName) {
  const xhttp = new XMLHttpRequest();
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
