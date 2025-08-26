{{/*
Expand the name of the chart.
*/}}
{{- define "app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "app.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := .Chart.Name }}
{{- if contains .Release.Name $name }}
{{- $name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: scheduler
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "app.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generage image reference based on image repository and tag
*/}}
{{- define "app.image" -}}
{{- printf "%s:%s" .Values.image.repository  (default (printf "%s" .Chart.AppVersion) .Values.image.tag) }}
{{- end }}

{{/*
Generage common labels
*/}}
{{- define "app.labels" -}}
helm.sh/chart: {{ include "app.chart" . }}
{{ include "app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.additionalLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Generate the Template editor cluster role name
*/}}
{{- define "app.clusteRole.templateEditor" -}}
{{- if .Values.rbac.clusterRole.templateEditor.nameOverride }}
{{- .Values.rbac.clusterRole.templateEditor.nameOverride }}
{{- else }}
{{- printf "%s-template-editor" (include "app.fullname" .) }}
{{- end }}
{{- end }}


{{/*
Generate the Workgroup editor cluster role name
*/}}
{{- define "app.clusteRole.workgroupEditor" -}}
{{- if .Values.rbac.clusterRole.workgroupEditor.nameOverride }}
{{- .Values.rbac.clusterRole.workgroupEditor.nameOverride }}
{{- else }}
{{- printf "%s-workgroup-editor" (include "app.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Generate the Scheduler cluster role name
*/}}
{{- define "app.clusteRole.scheduler" -}}
{{- if .Values.rbac.clusterRole.workgroupEditor.nameOverride }}
{{- .Values.rbac.clusterRole.workgroupEditor.nameOverride }}
{{- else }}
{{- printf "%s-scheduler" (include "app.fullname" .) }}
{{- end }}
{{- end }}
