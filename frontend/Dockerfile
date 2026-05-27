FROM oven/bun:1.3-alpine AS builder
# vue-tsc's #!/usr/bin/env node runs under bun's node-compat shim by default,
# where the Vue SFC TS plugin fails to activate ("Cannot find module './App.vue'").
# Installing real nodejs gives vue-tsc a working runtime.
RUN apk add --no-cache nodejs
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/templates/default.conf.template
ENV BACKEND_HOST=backend
EXPOSE 80
