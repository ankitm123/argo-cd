# Triggers and Templates Catalog
## Getting Started
* Install Triggers and Templates from the catalog
  ```bash
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/notifications_catalog/install.yaml
  ```
## Triggers
|          NAME          |                          DESCRIPTION                          |                      TEMPLATE                       |
|------------------------|---------------------------------------------------------------|-----------------------------------------------------|
| on-created             | Application is created.                                       | [app-created](#app-created)                         |
| on-deleted             | Application is deleted.                                       | [app-deleted](#app-deleted)                         |
| on-deployed            | Application is synced and healthy. Triggered once per commit. | [app-deployed](#app-deployed)                       |
| on-health-degraded     | Application has degraded                                      | [app-health-degraded](#app-health-degraded)         |
| on-sync-failed         | Application syncing has failed                                | [app-sync-failed](#app-sync-failed)                 |
| on-sync-running        | Application is being synced                                   | [app-sync-running](#app-sync-running)               |
| on-sync-status-unknown | Application status is 'Unknown'                               | [app-sync-status-unknown](#app-sync-status-unknown) |
| on-sync-succeeded      | Application syncing has succeeded                             | [app-sync-succeeded](#app-sync-succeeded)           |

## Templates
### app-created
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has been created.
message: Application {{.app.metadata.name}} has been created.
teams:
  title: Application {{.app.metadata.name}} has been created.

```
### app-deleted
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has been deleted.
message: Application {{.app.metadata.name}} has been deleted.
teams:
  title: Application {{.app.metadata.name}} has been deleted.

```
### app-deployed
**definition**:
```yaml
email:
  subject: New version of an application {{.app.metadata.name}} is up and running.
message: |
  {{if eq .serviceType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} is now running new version of deployments manifests.
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#18be52",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      },
      {
        "title": "Revision",
        "value": "{{.app.status.sync.revision}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Sync Status",
      "value": "{{.app.status.sync.status}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    },
    {
      "name": "Revision",
      "value": "{{.app.status.sync.revision}}"
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Operation Application",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
      }]
    }]
  themeColor: '#000080'
  title: New version of an application {{.app.metadata.name}} is up and running.

```
### app-health-degraded
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has degraded.
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} has degraded.
  Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#f4c030",
      "fields": [
      {
        "title": "Health Status",
        "value": "{{.app.status.health.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Health Status",
      "value": "{{.app.status.health.status}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Open Application",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
      }]
    }]
  themeColor: '#FF0000'
  title: Application {{.app.metadata.name}} has degraded.

```
### app-sync-failed
**definition**:
```yaml
email:
  subject: Failed to sync application {{.app.metadata.name}}.
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}}  The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#E96D76",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Sync Status",
      "value": "{{.app.status.sync.status}}"
    },
    {
      "name": "Failed at",
      "value": "{{.app.status.operationState.finishedAt}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Open Operation",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}{{ $source.repoURL }}⬆️ {{- end }}" {{- end }}
      }]
    }]
  themeColor: '#FF0000'
  title: Failed to sync application {{.app.metadata.name}}.

```
### app-sync-running
**definition**:
```yaml
email:
  subject: Start syncing application {{.app.metadata.name}}.
message: |
  The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#0DADEA",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Sync Status",
      "value": "{{.app.status.sync.status}}"
    },
    {
      "name": "Started at",
      "value": "{{.app.status.operationState.startedAt}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Open Operation",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
      }]
    }]
  title: Start syncing application {{.app.metadata.name}}.

```
### app-sync-status-unknown
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} sync status is 'Unknown'
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} sync is 'Unknown'.
  Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
  {{if ne .serviceType "slack"}}
  {{range $c := .app.status.conditions}}
      * {{$c.message}}
  {{end}}
  {{end}}
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#E96D76",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Sync Status",
      "value": "{{.app.status.sync.status}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Open Application",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
      }]
    }]
  title: Application {{.app.metadata.name}} sync status is 'Unknown'

```
### app-sync-succeeded
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has been successfully synced.
message: |
  {{if eq .serviceType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#18be52",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
        "value": {{- if .app.spec.source }} ":arrow_heading_up: {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}:arrow_heading_up: {{ $source.repoURL }}{{- end }}" {{- end }},
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      ,
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]
  deliveryPolicy: Post
  groupingKey: ""
  notifyBroadcast: false
teams:
  facts: |
    [{
      "name": "Sync Status",
      "value": "{{.app.status.sync.status}}"
    },
    {
      "name": "Synced at",
      "value": "{{.app.status.operationState.finishedAt}}"
    },
    {
      "name": {{- if .app.spec.source }} "Repository" {{- else if .app.spec.sources }} "Repositories" {{- end }},
      "value": {{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
    }
    {{range $index, $c := .app.status.conditions}}
      ,
      {
        "name": "{{$c.type}}",
        "value": "{{$c.message}}"
      }
    {{end}}
    ]
  potentialAction: |
    [{
      "@type":"OpenUri",
      "name":"Operation Details",
      "targets":[{
        "os":"default",
        "uri":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true"
      }]
    },
    {
      "@type":"OpenUri",
      "name":"Open Repository",
      "targets":[{
        "os":"default",
        "uri":{{- if .app.spec.source }} "⬆️ {{ .app.spec.source.repoURL }}" {{- else if .app.spec.sources }} "{{- range $index, $source := .app.spec.sources }}{{ if $index }}\n{{ end }}⬆️ {{ $source.repoURL }}{{- end }}" {{- end }}
      }]
    }]
  themeColor: '#000080'
  title: Application {{.app.metadata.name}} has been successfully synced

```
