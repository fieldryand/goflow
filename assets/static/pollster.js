function pollster(jobName) {
	function pollJobState() {
	  var xhttp = new XMLHttpRequest();
	  xhttp.onreadystatechange = function() {
	    if (this.readyState == 4 && this.status == 200) {
	      console.log(JSON.parse(this.responseText));
	      document.getElementById(jobName).innerHTML = JSON.parse(this.responseText).state;
	    }
	  };
	  xhttp.open("GET", `/jobs/${jobName}/state`, true);
	  xhttp.send();
	  setTimeout(pollJobState, 2000);
	}
	return pollJobState
}
