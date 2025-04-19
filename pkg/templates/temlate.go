package templates

import (
	"html/template"
)

func PrepareTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Keys</title>
</head>
<body>
    <h1>Keys:</h1>
    <ul>
        {{range .}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</body>
</html>
`
	return template.Must(template.New("list").Parse(tmpl))
}
