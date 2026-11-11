# 0006. Same origin through a reverse proxy

## Context

The SPA and the service are separate containers. Calling the service directly from the browser would require CORS configuration and environment-specific API URLs baked into the build.

## Decision

The browser only talks to one origin. In production nginx serves the static build and proxies `/api/` to `calculator-service`; in development the Vite dev server proxies `/api` to `localhost:8080`. The frontend uses relative URLs.

## Consequences

- No CORS policy to configure or secure.
- The same frontend build runs in any environment.
- The service is not exposed to the host in the compose setup; it is reachable only through the proxy.
