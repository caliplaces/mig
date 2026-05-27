// Package docs serves the OpenAPI spec and a Swagger UI page.
package docs

import _ "embed"

//go:embed openapi.yaml
var Spec []byte

// SwaggerUI is a self-contained HTML page that loads Swagger UI from a CDN
// and renders the embedded OpenAPI spec served at /openapi.yaml.
const SwaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>MIG API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.addEventListener('load', () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        layout: 'BaseLayout',
      });
    });
  </script>
</body>
</html>
`
