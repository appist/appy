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

				#type-toggle {
					display: flex !important;
					flex-basis: auto;
					flex-grow: 1;
					align-items: center;
				}

				#content {
					display: flex;
					flex-direction: column;
					min-width: 100vw;
				}

				#content > .container-fluid {
					display: flex;
					flex: 1;
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
				var previewURL = '{{.bbaseURL}}/preview';

				$("#menu-toggle").click(function(e) {
					e.preventDefault()
					$("#wrapper").toggleClass("toggled")
				})

				function setCurrPreview(targetName, targetExt) {
					var name = targetName || queryParam('name'), ext = targetExt || queryParam('ext') || 'html'

					if (name) {
						location.search = '?name=' + name + '&ext=' + ext
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

func previewTpl() string {
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

				<div id="type-toggle">
					<div class="btn-group btn-group-toggle ml-auto mt-lg-0" data-toggle="buttons">
						<label class="btn btn-primary{{if eq .ext "html"}} active{{end}}" onclick="onPreviewExtClicked(event, 'html')">
							<input type="radio" name="options" autocomplete="off"> HTML
						</label>
						<label class="btn btn-primary{{if eq .ext "txt"}} active{{end}}" onclick="onPreviewExtClicked(event, 'txt')">
							<input type="radio" name="options" autocomplete="off"> Text
						</label>
					</div>
				</div>
			</nav>

			<div class="container-fluid">
				{{if .name}}
					<iframe src="{{.baseURL}}/preview?name={{.name}}&ext={{.ext}}" frameBorder="0"></iframe>
				{{else}}
					Oops! Have you forgotten to setup the preview?
				{{end}}
			</div>
		</div>
  	</div>
` + tplLower()
}
