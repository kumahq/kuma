package gotemplates

var KriEndpointTemplate = `
openapi: 3.1.0
info:
  version: v1alpha1
  title: Kuma API
  description: Kuma API

paths:
  /_kri/{kri}:
    get:
      operationId: getByKri
      summary: Returns a resource by KRI
      tags: [ "KRI" ]
      parameters:
        - in: path
          name: kri
          schema:
            type: string
          required: true
          description: KRI of the resource
      responses:
        '200':
          description: The resource
          content:
            application/json:
              schema:
                oneOf:
{{- range .Resources }}
                  - $ref: '{{.Path}}#/components/schemas/{{.ResourceType}}Item'
{{- end }}
        '400':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
        '404':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/NotFound"
  /_kri/{kri}/_overview:
    get:
      operationId: getOverviewByKri
      summary: Returns the overview of a resource by KRI
      description: Returns the resource merged with its insight. Only resource types that have insights are supported.
      tags: [ "KRI" ]
      parameters:
        - in: path
          name: kri
          schema:
            type: string
          required: true
          description: KRI of the resource
      responses:
        '200':
          description: The resource overview
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '/specs/protoresources/dataplaneoverview/rest.yaml#/components/schemas/DataplaneOverviewWithMeta'
                  - $ref: '/specs/protoresources/meshoverview/rest.yaml#/components/schemas/MeshOverviewWithMeta'
        '400':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
        '404':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/NotFound"
`
