var dagreD3 = require("dagre-d3");
var d3 = require("d3");

async function getDag(jobName) {
  const response = await fetch(`/api/jobs/${jobName}`);
  const json = await response.json();
  return json.dag
}

export async function graphViz(jobName) {

  const dag = await getDag(jobName);

  // Create a new directed graph
  var g = new dagreD3.graphlib.Graph().setGraph({});

  for (var key in dag) {
    g.setNode(key, { id: "node-" + key, label: key });
    for (var val in dag[key]) {
      g.setEdge(key, dag[key][val], {});
    }
  }

  var svg = d3.select("svg");
  var inner = svg.select("g");

  g.nodes().forEach(function(v) {
    var node = g.node(v);
    node.rx = node.ry = 5;
  });
  
  // Set up zoom support
  var zoom = d3.zoom().on("zoom", function() {
    inner.attr("transform", d3.event.transform);
  });
  svg.call(zoom);
  
  // Create the renderer
  var render = new dagreD3.render();
  
  // Run the renderer. This is what draws the final graph.
  render(inner, g);
  
  // Center the graph
  var initialScale = 1.00;
  svg.call(zoom.transform, d3.zoomIdentity.translate(20, 20).scale(initialScale));
  
  svg.attr('height', g.graph().height * initialScale + 40);
}
