function pollster(jobName) {
	function pollJobState() {
	  var xhttp = new XMLHttpRequest();
	  xhttp.onreadystatechange = function() {
	    if (this.readyState == 4 && this.status == 200) {
	      var parsed = JSON.parse(this.response)
	      document.getElementById(jobName).innerHTML = JSON.stringify(parsed.jobRuns);
	    }
	  };
	  xhttp.open("GET", `/jobs/${jobName}/jobRuns`, true);
	  xhttp.send();
	  setTimeout(pollJobState, 2000);
	}
	return pollJobState
}
