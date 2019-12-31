package mailer

func tplUpper() string {
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

func tplLower() string {
	return `
			</main>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<script src="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/js/bootstrap.min.js"></script>
			<script>
				var previewURL = '{{.baseURL}}/preview'

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
		</body>
	</html>
`
}

// PreviewTpl returns the preview HTML page.
func PreviewTpl() string {
	return tplUpper() + `
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
							<iframe src="{{.baseURL}}/preview?name={{.name}}&ext={{.ext}}&locale={{.locale}}" frameBorder="0"></iframe>
						</div>
					</div>
				{{else}}
					Oops! Have you forgotten to setup the preview?
				{{end}}
			</div>
		</div>
  	</div>
` + tplLower()
}
