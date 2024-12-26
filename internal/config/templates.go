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
{{- $firstMaster := (index (filterMasterServers .ALLServers) 0) }}
                {{ $firstMaster.Name }}:
            masters-connect:
              hosts:
{{- range .ALLServers }}
{{- if and (eq (index .Roles 0) "control-plane") (ne .Name $firstMaster.Name) }}
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
