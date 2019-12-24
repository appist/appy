package http

import (
	"github.com/appist/appy/support"
)

func errorTplUpper() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
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

func errorTplLower() string {
	return `
			</main>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<script src="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/js/bootstrap.min.js"></script>
		</body>
	</html>
	`
}

func errorTpl404() string {
	return errorTplUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">The page that you are looking for does not exist, please contact the website administrator for more details.</p>
	</div>
</div>
		` + errorTplLower()
}

func errorTpl500() string {
	if support.IsDebugBuild() {
		return errorTplUpper() + `
<h2 class="text-danger">Full Trace</h2>
<pre class="pre-scrollable bg-light p-2">{{range $error := .errors}}{{$error}}{{end}}</pre>
<h2 class="text-danger">Request</h2>
<h6>Headers</h6>
<pre class="pre-scrollable bg-light p-2">{{.headers}}</pre>
<h6>Query String Parameters</h6>
<pre class="pre-scrollable bg-light p-2">{{.qsParams}}</pre>
<h6>Session Variables</h6>
<pre class="pre-scrollable bg-light p-2">{{.sessionVars}}</pre>
		` + errorTplLower()
	}

	return errorTplUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">If you are the administrator of this website, then please read this web application's log file and/or the web server's log file to find out what went wrong.</p>
	</div>
</div>
	` + errorTplLower()
}

func welcomeTpl() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<title>{{.title}}</title>
			<link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" />
			<style>
				/*
				* Base structure
				*/
				html,
				body {
					height: 100%;
				}
				body {
					display: -ms-flexbox;
					display: flex;
				}
				.cover-container {
					max-width: 48em;
				}
				/*
				* Header
				*/
				.masthead {
					margin-bottom: 2rem;
				}
				.masthead-brand {
					margin-bottom: 0;
				}
				@media (min-width: 48em) {
					.masthead-brand {
						float: left;
					}
				}
				/*
				* Cover
				*/
				.cover {
					margin-top: -3rem;
					padding: 0 1.5rem;
				}
				.cover .lead:first-of-type {
					margin-top: -3rem;
					margin-bottom: 1.5rem;
				}
				/*
				* Footer
				*/
				.mastfoot {
					color: rgba(255, 255, 255, .5);
				}
			</style>
		</head>
		<body class="text-center">
			<div class="cover-container d-flex w-100 h-100 p-2 mx-auto flex-column">
				<header class="masthead mb-auto"></header>
				<main role="main" class="inner cover">
					<h1 class="cover-heading">` + logoImage() + `</h1>
					<p class="lead">An opinionated productive web framework that helps scaling business easier.</p>
					<p class="lead">
						<a href="https://appy.appist.io" class="btn btn-lg btn-primary">Learn more</a>
					</p>
				</main>
				<footer class="mastfoot mt-auto"></footer>
			</div>
		</body>
	</html>
	`
}

func logoImage() string {
	return `
	<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="18rem" height="18rem" viewBox="0 0 1024 1024" xml:space="preserve">
		<g transform="matrix(1 0 0 1 512 512)" id="background-logo">
			<rect style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(255,255,255); fill-rule: nonzero; opacity: 1;"  paint-order="stroke"  x="-512" y="-512" rx="0" ry="0" width="1024" height="1024" />
		</g>
		<g transform="matrix(0.45660522273425497 0 0 -0.45660522273425497 512 395.2665130568356)" id="maker-logo">
			<path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(44,123,229); fill-rule: nonzero; opacity: 1;"  paint-order="stroke"  transform=" translate(-815.895, -767.6175000000001)" d="m 1082.67 409.402 l -55.05 -31.785 V 708.918 L 815.895 587.73 L 604.176 708.918 V 377.617 l -55.059 31.785 v 594.508 l 55.059 -31.785 V 757.188 L 815.895 634.949 L 1027.62 757.188 v 215.21 l 55.05 31.782 z M 841.746 864.5 L 975.906 941.957 V 787.043 l -134.16 -77.457 z m -51.707 0 V 709.586 l -134.156 77.457 v 154.914 z m -108.301 122.234 l 134.157 77.456 l 134.16 -77.456 l -134.16 -77.453 z m 134.157 200.736 l 240.415 -138.8 l -55.06 -31.78 l -185.355 107.01 l -185.594 -107.15 l -55.055 31.79 z m 0 59.71 L 497.41 1063.3 V 379.555 l 106.766 -61.645 l 51.707 -29.855 v 59.707 v 272.515 l 160.012 -92.386 l 160.011 92.386 V 347.766 V 288.059 l 51.714 29.855 l 106.76 61.641 V 1063.3 l -318.485 183.88" stroke-linecap="round" />
		</g>
		<g transform="matrix(1.5745007680491552 0 0 1.5745007680491552 512.0074881042386 703.06159621802)">
			<filter id="SVGID_14400" y="-20%" height="140%" x="-20%" width="140%">
				<feGaussianBlur in="SourceAlpha" stdDeviation="0"></feGaussianBlur>
				<feOffset dx="0" dy="0" result="oBlur" ></feOffset>
				<feFlood flood-color="rgb(0,0,0)" flood-opacity="1"/>
				<feComposite in2="oBlur" operator="in" />
				<feMerge>
					<feMergeNode></feMergeNode>
					<feMergeNode in="SourceGraphic"></feMergeNode>
				</feMerge>
			</filter>
			<path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(44,123,229); fill-rule: nonzero; opacity: 1;filter: url(#SVGID_14400);"  paint-order="stroke"  transform=" translate(-101.43999999999998, 12.774999999999999)" d="M 29.05 0 L 6.24 0 L 6.24 -22.71 L 34.63 -22.71 L 34.63 -31.24 L 17.61 -31.24 L 17.61 -28.39 L 6.24 -28.39 L 6.24 -34.43 L 6.24 -39.76 L 40.18 -39.76 Q 41.19 -38.83 45.94 -34.08 L 45.94 -34.08 L 45.94 0 L 34.63 0 L 34.63 -2.84 Q 33.21 -2.11 29.05 0 L 29.05 0 Z M 17.61 -11.72 L 17.61 -8.53 L 28.98 -8.53 Q 29.54 -8.81 34.63 -11.3 L 34.63 -11.3 L 34.63 -14.21 L 26.14 -14.21 L 17.61 -14.21 L 17.61 -11.72 Z M 58.48 14.21 L 58.48 -13.83 L 58.48 -39.76 L 64.17 -39.76 L 69.86 -39.76 L 69.86 -34.08 L 78.38 -39.76 L 98.18 -39.76 L 98.18 -18.44 L 98.18 -11.37 L 98.18 -5.69 L 96.13 -3.54 L 92.56 0 L 69.86 0 L 69.86 14.21 L 58.48 14.21 Z M 69.86 -17.06 L 69.86 -8.53 L 86.88 -8.53 L 86.88 -31.24 L 78.38 -31.24 Q 76.58 -29.95 74.05 -28.39 L 74.05 -28.39 Q 73.7 -28.19 72.2 -27.15 Q 70.69 -26.1 69.86 -25.55 L 69.86 -25.55 L 69.86 -17.06 Z M 110.73 14.21 L 110.73 -13.83 L 110.73 -39.76 L 116.42 -39.76 L 122.1 -39.76 L 122.1 -34.08 L 130.63 -39.76 L 150.42 -39.76 L 150.42 -18.44 L 150.42 -11.37 L 150.42 -5.69 L 148.38 -3.54 L 144.81 0 L 122.1 0 L 122.1 14.21 L 110.73 14.21 Z M 122.1 -17.06 L 122.1 -8.53 L 139.12 -8.53 L 139.12 -31.24 L 130.63 -31.24 Q 128.83 -29.95 126.3 -28.39 L 126.3 -28.39 Q 125.95 -28.19 124.44 -27.15 Q 122.93 -26.1 122.1 -25.55 L 122.1 -25.55 L 122.1 -17.06 Z M 185.27 -11.37 L 162.56 -11.37 L 162.56 -25.55 L 162.56 -39.76 L 173.93 -39.76 L 173.93 -34.08 L 173.93 -22.71 L 185.27 -22.71 L 185.27 -39.76 L 196.64 -39.76 L 196.64 8.53 L 168.24 8.53 L 168.24 0 L 185.27 0 L 185.27 -11.37 Z" stroke-linecap="round" />
		</g>
	</svg>
`
}
