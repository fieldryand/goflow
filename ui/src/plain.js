function getDropdownValue() {
  const selectDropdown = document.querySelector('select');
  return selectDropdown.value
}

function indexPageEventListener() {
  var stream = new EventSource(`/stream`);
  stream.addEventListener("message", indexPageEventHandler)
}

function jobPageEventListener(job) {
  var stream = new EventSource(`/stream?jobname=${job}`);
  stream.addEventListener("message", jobPageEventHandler)
}

function indexPageEventHandler(message) {
  const d = JSON.parse(message.data);
  const jobID = d.id;
  const s = stateColor(d.state);
  const ts = d.submitted;
  updateStateCircles("job-table", jobID, d.job, s, ts);
}

function jobPageEventHandler(message) {
  const d = JSON.parse(message.data);
  updateTaskStateCircles(d);
  updateGraphViz(d);
  updateLastRunTs(d);
}

function updateStateCircles(tableName, jobID, wrapperId, color, startTimestamps) {
  const wrapper = document.getElementById(wrapperId);
  const startTimestamp = startTimestamps;
  div = document.createElement("div");
  div.setAttribute("id", jobID);
  div.setAttribute("class", "status-indicator");
  div.setAttribute("style", `background-color:${color}`);
  div.setAttribute("title", jobID);
  if (jobID in wrapper.children) {
    wrapper.replaceChild(div, document.getElementById(jobID));
  } else {
    wrapper.appendChild(div);
  }
}

function updateTaskStateCircles(execution) {
  var tasks = {};
  var startTimestamps = {};

  const taskList = execution.tasks;
  const startTimestamp = execution.submitted;

  for (j in taskList) {
    const state = taskList[j].state;
    const taskName = taskList[j].name;
    const color = stateColor(state);
    if (taskName in tasks) {
      tasks[taskName].push(color);
      startTimestamps[taskName].push(startTimestamp);
    } else {
      tasks[taskName] = [color];
      startTimestamps[taskName] = [startTimestamp];
    }
  }

  for (task in tasks) {
    updateStateCircles("task-table", `${execution.id}-${task}`, task, tasks[task], startTimestamps[task]);
  }
}

function updateGraphViz(execution) {
  const tasks = execution.tasks
  for (i in tasks) {
    if (document.getElementsByClassName("output")) {
      try {
        const rect = document.getElementById("node-" + tasks[i].name).querySelector("rect");
        rect.setAttribute("style", "stroke-width: 2; stroke: " + stateColor(tasks[i].state));
      }
      catch(err) {
        console.log(`${err}. This might be a temporary error when the graph is still loading.`)
      }
    }
  }
}

function updateLastRunTs(execution) {
  const lastExecutionTs = execution.submitted;
  const lastExecutionTsHTML = document.getElementById("last-execution-ts-wrapper").innerHTML;
  const newHTML = lastExecutionTsHTML.replace(/.*/, `Last run: ${lastExecutionTs}`);
  document.getElementById("last-execution-ts-wrapper").innerHTML = newHTML;
}

function updateJobActive(jobName) {
  fetch(`/api/jobs/${jobName}`)
    .then(response => response.json())
    .then(data => {
      if (data.active) {
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

function lastN(array, n) {
  reversed = array.toReversed()
  sliced = reversed.slice(0, n)
  return sliced.toReversed()
}


function stateColor(taskState) {
  switch (taskState) {
    case "running":
      var color = "#dffbe3";
      break;
    case "upforretry":
      var color = "#ffc620";
      break;
    case "successful":
      var color = "#39c84e";
      break;
    case "skipped":
      var color = "#abbefb";
      break;
    case "failed":
      var color = "#ff4020";
      break;
    case "notstarted":
      var color = "white";
      break;
  }

  return color
}

async function buttonPress(buttonName, jobName) {
  var button = document.getElementById(`button-${buttonName}-${jobName}`);
  // Add 'clicked' class to apply the style
  button.classList.add('clicked');

  const options = {
    method: 'POST'
  }
  await fetch(`/api/jobs/${jobName}/${buttonName}`, options)
    .then(updateJobActive(jobName))

  setTimeout(function() {
    button.classList.remove('clicked');
  }, 200); // 200 milliseconds delay
}
