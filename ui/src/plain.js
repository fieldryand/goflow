function indexPageEventListener() {
  const dateControl = document.querySelector('input[type="date"]');
  if (dateControl.value == "") {
    var date = today();
  } else {
    var date = dateControl.value;
  }
  var stream = new EventSource(`/events?date=${date}`);
  stream.addEventListener("message", indexPageEventHandler)
}

function jobPageEventListener(job) {
  const dateControl = document.querySelector('input[type="date"]');
  if (dateControl.value == "") {
    var date = today();
  } else {
    var date = dateControl.value;
  }
  var stream = new EventSource(`/events/${job}?date=${date}`);
  stream.addEventListener("message", jobPageEventHandler)
}

function diagramPageEventListener(job) {
  const dateControl = document.querySelector('input[type="date"]');
  if (dateControl.value == "") {
    var date = today();
  } else {
    var date = dateControl.value;
  }
  var stream = new EventSource(`/events/${job}?date=${date}`);
  stream.addEventListener("message", diagramPageEventHandler)
}

function today() {
  let today = new Date().toISOString().slice(0, 10);
  return today
}

function indexPageEventHandler(message) {
  const d = JSON.parse(message.data);
  const s = stateColor(d.state);
  updateStateCircles("job-table", d.id, d.job, s, d.startTs);
  updateLastStart(d);
}

function jobPageEventHandler(message) {
  const d = JSON.parse(message.data);
  updateTaskStateCircles(d);
  updateLastTaskStartModified(d);
}

function diagramPageEventHandler(message) {
  const d = JSON.parse(message.data);
  updateLastRunTs(d);
  updateGraphViz(d);
}

function updateStateCircles(tableName, jobID, wrapperId, color, startTimestamp) {
  const options = {
    dateStyle: 'medium',
    timeStyle: 'medium'
  };
  const wrapper = document.getElementById(wrapperId);
  const startTs = new Date(startTimestamp);
  const formattedTs = startTs.toLocaleString(undefined, options); 
  div = document.createElement("div");
  div.setAttribute("id", jobID);
  div.setAttribute("class", "status-indicator");
  div.setAttribute("style", `background-color:${color}`);
  div.setAttribute("title", `ID: ${jobID}\nStarted: ${formattedTs}`);
  if (jobID in wrapper.children) {
    wrapper.replaceChild(div, document.getElementById(jobID));
  } else {
    wrapper.appendChild(div);
  }
}

function updateTaskStateCircles(execution) {
  for (i in execution.tasks) {
    const t = execution.tasks[i];
    const s = stateColor(t.state);
    updateStateCircles("task-table", `${execution.id}-${t.name}`, t.name, s, execution.startTs);
  }
}

function updateLastStart(execution) {
  const options = {
    dateStyle: 'medium',
    timeStyle: 'medium'
  };
  const startTs = new Date(execution.startTs);
  const formattedTs = startTs.toLocaleString(undefined, options);
  const job = execution.job;
  document.getElementById(`last-start-${job}`).innerHTML = formattedTs;
}

function updateLastTaskStartModified(execution) {
  const options = {
    dateStyle: 'medium',
    timeStyle: 'medium'
  };
  for (i in execution.tasks) {
    const t = execution.tasks[i];
    const startTs = new Date(t.startTs);
    const modifiedTs = new Date(t.modifiedTs);
    // check that both timestamps are not in year 1 (the 0-value)
    if (startTs.getUTCFullYear() > 1 & modifiedTs.getUTCFullYear() > 1) {
      const seconds = (modifiedTs - startTs) / 1000;
      const duration = new Date(0);
      duration.setSeconds(seconds);
      const durationStr = duration.toISOString().substring(11, 19);
      const formattedStartTs = startTs.toLocaleString(undefined, options);
      document.getElementById(`last-start-${t.name}`).innerHTML = formattedStartTs;
      document.getElementById(`last-duration-${t.name}`).innerHTML = durationStr;
    }
  }
}

function updateGraphViz(execution) {
  const tasks = execution.tasks
  for (i in tasks) {
    if (document.getElementsByClassName("output")) {
      try {
        const taskElem = document.querySelector("[data-id=\""+tasks[i].name+"\"]");
	let rect = taskElem.querySelector("rect");
        rect.setAttribute("style", "stroke-width: 2; stroke: " + stateColor(tasks[i].state));
      }
      catch(err) {
        console.log(`${err}. This might be a temporary error when the graph is still loading.`)
      }
    }
  }
}

function updateLastRunTs(execution) {
  const options = {
    dateStyle: 'medium',
    timeStyle: 'medium'
  };
  const startTs = new Date(execution.startTs);
  const formattedTs = startTs.toLocaleString(undefined, options);
  const lastExecutionTsHTML = document.getElementById("last-execution-ts-wrapper").innerHTML;
  const newHTML = lastExecutionTsHTML.replace(/.*/, `Last run: ${formattedTs}`);
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
