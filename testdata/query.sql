SELECT * FROM table
WHERE {{ if .Since }}updatetime > {{ .Since }}{{ else }}TRUE{{ end }}
