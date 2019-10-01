package html

import "github.com/appist/appy/support"

func errorUpper() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
	  	<head>
			<title>{{.title}}</title>
			<link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" />
			<style>
			body { padding-top: 4.5rem; }
			</style>
	  	</head>
		<body>
			<nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark">
				<div class="navbar-brand">{{.title}}</div>
			</nav>
			<main role="main" class="px-3">
	`
}

func errorLower() string {
	return `
			</main>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<script src="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/js/bootstrap.min.js"></script>
		</body>
	</html>
	`
}

// ErrorTpl404 returns the template for 404 HTTP error.
func ErrorTpl404() string {
	return errorUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">The page that you are looking for does not exist, please contact the website administrator for more details.</p>
	</div>
</div>
		` + errorLower()
}

// ErrorTpl500 returns the template for 500 HTTP error.
func ErrorTpl500() string {
	if support.Build == "debug" {
		return errorUpper() + `
<h2 class="text-danger">Full Trace</h2>
<pre class="pre-scrollable bg-light p-2">{{range $error := .errors}}{{$error}}{{end}}</pre>
<h2 class="text-danger">Request</h2>
<h6>Headers</h6>
<pre class="pre-scrollable bg-light p-2">{{.headers}}</pre>
<h6>Query String Parameters</h6>
<pre class="pre-scrollable bg-light p-2">{{.qsParams}}</pre>
<h6>Session Variables</h6>
<pre class="pre-scrollable bg-light p-2">{{.sessionVars}}</pre>
		` + errorLower()
	}

	return errorUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">If you are the administrator of this website, then please read this web application's log file and/or the web server's log file to find out what went wrong.</p>
	</div>
</div>
	` + errorLower()
}
