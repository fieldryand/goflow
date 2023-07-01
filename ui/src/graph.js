import mermaid from "mermaid";

async function getDag(jobName) {
  const response = await fetch(`/api/jobs/${jobName}`);
  const json = await response.json();
  return json.dag;
}

export async function graphViz(jobName) {
  const dag = await getDag(jobName);

  mermaid.initialize({ startOnLoad: false });

  var graphDefinition = "graph LR;\n";

  for (var key in dag) {
    graphDefinition += `  ${key}["${key}"];\n`;
    for (var val in dag[key]) {
      graphDefinition += `  ${key} --> ${dag[key][val]};\n`;
    }
  }

  console.log(graphDefinition);

  mermaid.render("graph", graphDefinition, function(svgCode) {
    var svgContainer = document.querySelector("#graph-container");
    svgContainer.innerHTML = svgCode;
  });
}
