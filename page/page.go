package page

import (
	"net/http"

	"github.com/jansemmelink/log"
)

//Start ...
func Start(res http.ResponseWriter, title string) {
	res.Write([]byte(`<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<!-- The above 3 meta tags *must* come first in the head; any other head content must come *after* these tags -->

			<title>` + title + `</title>

			<!-- Bootstrap -->
			<link href="/resource/css/bootstrap.min.css" rel="stylesheet">

			<!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
			<!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
			<!--[if lt IE 9]>
				<script src="https://oss.maxcdn.com/html5shiv/3.7.3/html5shiv.min.js"></script>
				<script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
			<![endif]-->
		</head>
		<body>

		<!-- MAIN MENU -->
		<nav class="navbar navbar-inverse">
		<div class="container-fluid">
		  <div class="navbar-header">
			<a class="navbar-brand" href="#">Get it done!</a>
		  </div>
		  <ul class="nav navbar-nav">
			<li class="active"><a href="/enter">Enter</a></li>
			<li><a href="/person/new">+Mens</a></li>
			<li><a href="/entries">Entries</a></li>
			<li><a href="/orgs">Organisasies</a></li>
			<li><a href="/persons">Mense</a></li>
			<li><a href="/auth/logout">Logout</a></li>
		  </ul>
		  <!--button class="btn btn-danger navbar-btn">Restart</button-->
		</div>
	  </nav>
	`))
}

//End ...
func End(res http.ResponseWriter) {
	res.Write([]byte(`
			<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
			<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
			<!-- Include all compiled plugins (below), or include individual files as needed -->
			<script src="/resource/js/bootstrap.min.js"></script>
		</body>
	</html>
	`))
}

//Error ...
func Error(res http.ResponseWriter, req *http.Request, err error) {
	log.Debugf("showing error: %v", err)
	Start(res, "Error")
	res.Write([]byte(`
	<h1>Error</h1>
	<p>Sorry. Something went wrong.</p>
	<p>Click <a href="/enter">here</a> to go home.</p>
	<p>Details: ` + err.Error() + `</p>
	`))
	End(res)
	log.Debugf("showed error: %v", err)
}
