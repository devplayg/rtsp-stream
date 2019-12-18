package ui

import (
	"fmt"
	"time"
)

func Base() string {
	return `<!doctype html>
<html lang="en">
 <body>
    {{ block "content" . }}{{ end }}
  <script src="https://cdn.jsdelivr.net/npm/vue"></script>
{{ block "script" . }}{{ end }}

 </body>
</html>
`
}

func Layout(body string) string {
	tpl := `<!doctype html>
<html>
<head>
	<meta charset="utf-8">
</head>
<body>
%s

{{with .Account -}}
Dear {{.FirstName}} {{.LastName}},
{{- end}}

Below are your account statement details for period from {{.FromDate | formatAsDate}} to {{.ToDate | formatAsDate}}.

{{if .Purchases -}}
    Your purchases:
    {{- range .Purchases }}
        {{ .Date | formatAsDate}} {{ printf "%-20s" .Description }} {{.AmountInCents | formatAsDollars -}}
    {{- end}}
{{- else}}
You didn't make any purchases during the period.
{{- end}}

{{$note := urgentNote .Account -}}
{{if $note -}}
Note: {{$note}}
{{- end}}

Best Wishes,
Customer Service

</body>
</html>`

	return fmt.Sprintf(tpl, body)
}

func Hello() string {
	return `{{define "title"}}A templated page{{end}}

{{define "body"}}
    <h1>Hello from a templated page</h1>
{{end}}`
}

type Account struct {
	FirstName string
	LastName  string
}

type Purchase struct {
	Date          time.Time
	Description   string
	AmountInCents int
}

type Statement struct {
	FromDate  time.Time
	ToDate    time.Time
	Account   Account
	Purchases []Purchase
}

func CreateMockStatement() Statement {
	return Statement{
		FromDate: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC),
		Account: Account{
			FirstName: "John",
			LastName:  "Dow",
		},
		Purchases: []Purchase{
			Purchase{
				Date:          time.Date(2016, 1, 3, 0, 0, 0, 0, time.UTC),
				Description:   "Shovel",
				AmountInCents: 2326,
			},
			Purchase{
				Date:          time.Date(2016, 1, 8, 0, 0, 0, 0, time.UTC),
				Description:   "Staple remover",
				AmountInCents: 5432,
			},
		},
	}
}
