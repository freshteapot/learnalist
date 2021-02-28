package api

import (
	"html/template"
)

//vars := make(map[string]interface{})
//vars["token"] = session.Token
//vars["userUUID"] = userUUID
//vars["refreshRedirectURL"] = "/welcome.html"
//vars["idp"] = user.IDPKeyApple
//
//var tpl bytes.Buffer
//oauthCallbackHtml200.Execute(&tpl, vars)

var oauthCallbackHtml200 = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head
	data-redirectUri="{{.refreshRedirectURL}}"
	data-token="{{.token}}"
	data-user-uuid="{{.userUUID}}"
>
<meta http-equiv="refresh" content="2;url={{.refreshRedirectURL}}" />

<meta charset="utf-8" />
<script>
const token = document.querySelector("head").getAttribute('data-token').toString();
const userUUID = document.querySelector("head").getAttribute('data-user-uuid').toString();
localStorage.setItem("app.user.authentication", JSON.stringify(token))
localStorage.setItem("app.user.uuid", JSON.stringify(userUUID))
</script>
</head>
<body>
<h1>You have successfully logged in via {{.idp}}</h1>
<p>You will now be redirected</p>
</body>
</html>
`))
