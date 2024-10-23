import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
let config = { startOnLoad: true, flowchart: { useMaxWidth: false, htmlLabels: true } };
mermaid.initialize(config);

async function getDag(jobName) {
  const response = await fetch(`/api/jobs/${jobName}`);
  const json = await response.json();
  return json.dag;
}

async function drawGraph() {
  var element = document.querySelector("div.graph-container");
  const jobName = element.dataset.job;
  const dag = await getDag(jobName);

  var graphDefinition = "graph LR;\n";

  for (var key in dag) {
    graphDefinition += `  ${key}["${key}"];\n`;
    for (var val in dag[key]) {
      graphDefinition += `  ${key} --> ${dag[key][val]};\n`;
    }
  }

  const { svg } = await mermaid.render("graph", graphDefinition);
  element.innerHTML = svg;
};

drawGraph();
