function updateStateCircles(tableName, wrapperId, colorArray, startTimestamps) {
  const oldWrapper = document.getElementById(wrapperId);
  const newWrapper = document.createElement("div");
  newWrapper.setAttribute("class", "status-wrapper");
  newWrapper.setAttribute("id", wrapperId);
  for (k in colorArray) {
    const color = colorArray[k];
    const startTimestamp = startTimestamps[k];
    div = document.createElement("div");
    div.setAttribute("class", "status-indicator");
    div.setAttribute("style", `background-color:${color}`);
    div.setAttribute("title", startTimestamp);
    newWrapper.appendChild(div);
  }
  document.getElementById(tableName).replaceChild(newWrapper, oldWrapper);
}

function updateTaskStateCircles(executions) {
  var tasks = {};
  var startTimestamps = {};
  for (i in executions) {
    const taskList = executions[i].tasks;
    //const startTimestamp = executions[i].startTimestamp;
    const startTimestamp = executions[i].submitted;
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
  }
  for (task in tasks) {
    updateStateCircles("task-table", task, tasks[task], startTimestamps[task]);
  }
}

function getDropdownValue() {
  const selectDropdown = document.querySelector('select');
  return selectDropdown.value
}

function updateJobStateCircles() {
  var stream = new EventSource(`/stream`);
  stream.addEventListener("message", function(e) {
    const n = getDropdownValue();
    const d = JSON.parse(e.data);
    const s = lastN(d.executions, n).map(x => stateColor(x.state));
    //const ts = d.executions.map(x => x.startTimestamp);
    const ts = d.executions.map(x => x.submitted);
    updateStateCircles("job-table", d.jobName, s, ts);
  });
}

function updateGraphViz(executions) {
  if (executions.length) {
    const tasks = executions.reverse()[0].tasks
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
}

function updateLastRunTs(executions) {
  if (executions.reverse()[0]) {
    //const lastExecutionTs = executions.reverse()[0].startTimestamp;
    const lastExecutionTs = executions.reverse()[0].submitted;
    const lastExecutionTsHTML = document.getElementById("last-execution-ts-wrapper").innerHTML;
    const newHTML = lastExecutionTsHTML.replace(/.*/, `Last run: ${lastExecutionTs}`);
    document.getElementById("last-execution-ts-wrapper").innerHTML = newHTML;
  }
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

function readTaskStream(jobName) {

  var stream = new EventSource(`/stream`);
  stream.addEventListener("message", function(e) {

    // n = display last n executions
    const n = getDropdownValue();
    const d = JSON.parse(e.data);

    if (jobName == d.jobName) {
      updateTaskStateCircles(lastN(d.executions, n));
      updateGraphViz(d.executions);
      updateLastRunTs(d.executions);
    }

  });
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
