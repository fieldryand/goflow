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
  const s = stateColor(d.state);
  updateStateCircles("job-table", d.id, d.job, s, d.submitted);
}

function jobPageEventHandler(message) {
  const d = JSON.parse(message.data);
  updateTaskStateCircles(d);
  updateGraphViz(d);
  updateLastRunTs(d);
}

function updateStateCircles(tableName, jobID, wrapperId, color, startTimestamp) {
  const limit = getDropdownValue();
  const wrapper = document.getElementById(wrapperId);
  div = document.createElement("div");
  div.setAttribute("id", jobID);
  div.setAttribute("class", "status-indicator");
  div.setAttribute("style", `background-color:${color}`);
  div.setAttribute("title", jobID);
  if (jobID in wrapper.children) {
    wrapper.replaceChild(div, document.getElementById(jobID));
  } else {
    if (wrapper.childElementCount >= limit) {
      wrapper.removeChild(wrapper.firstElementChild);
      wrapper.appendChild(div);
    } else {
      wrapper.appendChild(div);
    }
  }
}

function updateTaskStateCircles(execution) {
  for (i in execution.tasks) {
    const t = execution.tasks[i];
    const s = stateColor(t.state);
    updateStateCircles("task-table", `${execution.id}-${t.name}`, t.name, s, execution.submitted);
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
