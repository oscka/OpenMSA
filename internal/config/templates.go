package config

const (
        HostsTemplate = `127.0.0.1   localhost
{{- range .ALLServers }}
{{ .IP }} {{ .Name }}
{{- end }}
`

        AnsibleHostsTemplate = `all:
  children:
    control:
      hosts:
{{- range .ALLServers }}
{{- if eq (index .Roles 0) "control" }}
        {{ .Name }}:
{{- end }}
{{- end }}
    k8s-cluster:
      children:
        masters:
          children:
            master-init:
              hosts:
{{- range .ALLServers }}
{{- if and (eq (index .Roles 0) "control-plane") (eq .Name "rke2-master-node01") }}
                {{ .Name }}:
{{- end }}
{{- end }}
            masters-connect:
              hosts:
{{- range .ALLServers }}
{{- if and (eq (index .Roles 0) "control-plane") (ne .Name "rke2-master-node01") }}
                {{ .Name }}:
{{- end }}
{{- end }}
        workers:
          children:
            workers-group1:
              hosts:
{{- range .ALLServers }}
{{- if eq (index .Roles 0) "worker" }}
                {{ .Name }}:
{{- end }}
{{- end }}
`
)
