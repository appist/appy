package pack

import (
	"crypto/tls"
	"strings"

	"github.com/appist/appy/support"
)

const (
	// LiveReloadWSPort is the websocket port for the SSR live reload server.
	LiveReloadWSPort = "12450"

	// LiveReloadWSSPort is the websocket SSL port for the SSR live reload
	// server.
	LiveReloadWSSPort = "12451"

	// LiveReloadPath is the websocket path for the SSR live reload server.
	LiveReloadPath = "/reload"
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

func gqlPlaygroundTpl(path string, c *Context) []byte {
	return []byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset=utf-8/>
	<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
	<title>GraphQL Playground</title>
	<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/css/index.css" />
	<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/favicon.png" />
	<script src="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/js/middleware.js"></script>
</head>
<body>
	<div id="root">
	<style>
		body { background-color: rgb(23, 42, 58); font-family: Open Sans, sans-serif; height: 90vh; }
		#root { height: 100%; width: 100%; display: flex; align-items: center; justify-content: center; }
		.loading { font-size: 32px; font-weight: 200; color: rgba(255, 255, 255, .6); margin-left: 20px; }
		img { width: 78px; height: 78px; }
		.title { font-weight: 400; }
	</style>
	<img src="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/logo.png" alt="">
	<div class="loading"> Loading
		<span class="title">GraphQL Playground</span>
	</div>
	</div>
	<script>
		function getCookie(name) {
			var v = document.cookie.match('(^|;) ?' + name + '=([^;]*)(;|$)');
			return v ? v[2] : null;
		}
		window.addEventListener('load', function (event) {
			GraphQLPlayground.init(document.getElementById('root'), {
				endpoint: '` + path + `',
				subscriptionEndpoint: '` + path + `',
				headers: {
					'X-CSRF-Token': unescape(getCookie("` + mdwCSRFAuthenticityTemplateFieldName(c) + `"))
				},
				settings: {
					'request.credentials': 'include',
					'schema.polling.interval': 5000
				}
			})
		})
	</script>
</body>
</html>
`)
}

func liveReloadTpl(host string, isTLS *tls.ConnectionState) string {
	protocol := "ws"
	port := LiveReloadWSPort

	if isTLS != nil {
		protocol = "wss"
		port = LiveReloadWSSPort
	}

	splits := strings.Split(host, ":")
	url := protocol + `://` + splits[0] + ":" + port + LiveReloadPath

	return `<script>function b(a){var c=new WebSocket(a);c.onclose=function(){setTimeout(function(){b(a)},2E3)};` +
		`c.onmessage=function(){location.reload()}}try{if(window.WebSocket)try{b("` + url + `")}catch(a){console.error(a)}` +
		`else console.log("Your browser does not support WebSocket.")}catch(a){console.error(a)};</script>`
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

func mailerPreviewTpl() string {
	return mailerPreviewTplUpper() + `
	<div class="d-flex" id="wrapper">
    	<div class="bg-white border-right" id="sidebar">
			<div class="sidebar-heading bg-light">{{.title}}</div>
			<div class="list-group list-group-flush">
				{{range $idx, $preview := .previews}}
					<a
						href="#"
						class="list-group-item list-group-item-action{{if eq $.name $preview.Template}} list-group-item-dark{{end}}"
						onclick="onPreviewNameClicked(event)"
						data-name="{{$preview.Template}}">{{$preview.Template}}</a>
				{{end}}
			</div>
    	</div>
		<div id="content">
			<nav class="navbar navbar-expand-lg navbar-light bg-light border-bottom">
				<button class="btn" id="menu-toggle">
					<span class="navbar-toggler-icon"></span>
				</button>
			</nav>
			<div class="container-fluid p-3">
				{{if .name}}
					<div class="card">
						<div class="card-body row">
							<div class="col-auto">
								<table class="table table-borderless table-sm">
									<tbody>
										<tr>
											<th class="pr-4" scope="row">Subject</th>
											<td>{{.mail.Subject}}</td>
										</tr>
										<tr>
											<th class="pr-4" scope="row">From</th>
											<td>{{.mail.From}}</td>
										</tr>
										<tr>
											<th class="pr-4" scope="row">To</th>
											<td>
												{{range $idx, $val := .mail.To}}{{if $idx}}, {{end}}{{$val}}{{end}}
											</td>
										</tr>
										<tr>
											<th class="pr-4" scope="row">Reply To</th>
											<td>
												{{range $idx, $val := .mail.ReplyTo}}{{if $idx}}, {{end}}{{$val}}{{end}}
											</td>
										</tr>
										<tr>
											<th class="pr-4" scope="row">Cc</th>
											<td>
												{{range $idx, $val := .mail.Cc}}{{if $idx}}, {{end}}{{$val}}{{end}}
											</td>
										</tr>
										<tr>
											<th class="pr-4" scope="row">Bcc</th>
											<td>
												{{range $idx, $val := .mail.Bcc}}{{if $idx}}, {{end}}{{$val}}{{end}}
											</td>
										</tr>
									</tbody>
								</table>
							</div>
							<div class="col">
								<div class="toggle">
									<div class="btn-group btn-group-toggle ml-auto mt-lg-0" data-toggle="buttons">
										<button class="btn btn-primary{{if eq .ext "html"}} active{{end}}" onclick="onPreviewExtClicked(event, 'html')">
											<input type="radio" name="options" autocomplete="off"> HTML
										</button>
										<button class="btn btn-primary{{if eq .ext "txt"}} active{{end}}" onclick="onPreviewExtClicked(event, 'txt')">
											<input type="radio" name="options" autocomplete="off"> Text
										</button>
									</div>
								</div>
								<div class="toggle">
									<select class="custom-select ml-auto mt-lg-0" onchange="onPreviewLocaleChanged(event)">
										{{range $key, $val := .locales}}<option value="{{$val}}"{{if eq $.locale $val}} selected="true"{{end}}>{{$val}}</option>{{end}}
									</select>
								</div>
							</div>
						</div>
					</div>
					<div id="iframe-card" class="card mt-3">
						<div class="card-body">
							<iframe src="{{.path}}/preview?name={{.name}}&ext={{.ext}}&locale={{.locale}}" frameBorder="0"></iframe>
						</div>
					</div>
				{{else}}
					Oops! Have you forgotten to setup the preview?
				{{end}}
			</div>
		</div>
  	</div>
` + mailerPreviewTplLower()
}

func mailerPreviewTplUpper() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<title>{{.title}}</title>
			<link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" />
			<style>
				body {
					overflow-x: hidden;
				}
				iframe {
					width: 100%;
					height: 100%;
				}
				#sidebar {
					min-height: 100vh;
					margin-left: -15rem;
					-webkit-transition: margin .25s ease-out;
					-moz-transition: margin .25s ease-out;
					-o-transition: margin .25s ease-out;
					transition: margin .25s ease-out;
				}
				#sidebar .sidebar-heading {
					padding: 0.875rem 1.25rem;
					font-size: 1.2rem;
				}
				#sidebar .list-group {
					width: 15rem;
				}
				.toggle {
					display: flex !important;
					flex-basis: auto;
					flex-grow: 1;
					align-items: center;
				}
				.toggle > .btn-group, .toggle > .custom-select {
					width: 12rem;
					margin-bottom: 1rem;
				}
				#content {
					display: flex;
					flex-direction: column;
					min-width: 100vw;
					background-color: #F2F4F6;
				}
				#content > .container-fluid {
					display: flex;
					flex: 1;
					flex-direction: column;
					padding: 0;
				}
				#iframe-card {
					flex: 1;
				}
				#iframe-card > .card-body {
					padding: 0;
				}
				#wrapper.toggled #sidebar {
					margin-left: 0;
				}
				@media (min-width: 768px) {
					#sidebar {
						margin-left: 0;
					}
					#content {
						min-width: 0;
						width: 100%;
					}
					#wrapper.toggled #sidebar {
						margin-left: -15rem;
					}
				}
			</style>
	  	</head>
		<body>
`
}

func mailerPreviewTplLower() string {
	return `
			</main>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<script src="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/js/bootstrap.min.js"></script>
			<script>
				var previewURL = '{{.path}}/preview'
				$("#menu-toggle").click(function(e) {
					e.preventDefault()
					$("#wrapper").toggleClass("toggled")
				})
				document.addEventListener('DOMContentLoaded', function() {
					var name = queryParam('name') || '{{.name}}',
							ext = queryParam('ext') || '{{.ext}}',
							locale = queryParam('locale') || '{{.locale}}'
					history.replaceState('', '', '?name=' + name + '&ext=' + ext + '&locale=' + locale)
				})
				function setCurrPreview(targetName, targetExt, targetLocale) {
					var name = targetName || queryParam('name'),
							ext = targetExt || queryParam('ext') || 'html',
							locale = targetLocale || queryParam('locale')
					if (name) {
						location.search = '?name=' + name + '&ext=' + ext + '&locale=' + locale
					}
				}
				function onPreviewNameClicked(e) {
					e.preventDefault()
					setCurrPreview(e.target.dataset.name)
				}
				function onPreviewExtClicked(e, ext) {
					e.preventDefault()
					setCurrPreview(null, ext)
				}
				function onPreviewLocaleChanged(e) {
					e.preventDefault()
					setCurrPreview(null, null, e.target.value)
				}
				function queryParam(name) {
					var result = null, tmp = [];
					location.search
							.substr(1)
							.split("&")
							.forEach(function (item) {
								tmp = item.split("=")
								if (tmp[0] === name) result = decodeURIComponent(tmp[1])
							})
					return result
				}
			</script>
			{{.liveReloadTpl}}
		</body>
	</html>
`
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
