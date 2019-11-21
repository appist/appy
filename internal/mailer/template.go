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
					min-width: 100vw;
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
				$("#menu-toggle").click(function(e) {
					e.preventDefault();
					$("#wrapper").toggleClass("toggled");
				});
			</script>
		</body>
	</html>
`
}

func previewTpl() string {
	return tplUpper() + `
	<div class="d-flex" id="wrapper">
    	<div class="bg-light border-right" id="sidebar">
			<div class="sidebar-heading">{{.title}}</div>
			<div class="list-group list-group-flush">
				{{range $idx, $preview := .previews}}
					<a href="#{{$preview.Template}}" class="list-group-item list-group-item-action bg-light">{{$preview.Template}}</a>
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
						<label class="btn btn-primary">
							<input type="radio" name="options" id="html" autocomplete="off"> HTML
						</label>
						<label class="btn btn-primary">
							<input type="radio" name="options" id="text" autocomplete="off"> Text
						</label>
					</div>
				</div>
			</nav>

			<div class="container-fluid">
				<iframe src="/abc" frameBorder="0"></iframe>
			</div>
		</div>
  	</div>
` + tplLower()
}
