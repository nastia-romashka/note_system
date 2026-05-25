FROM node:22-alpine AS builder

WORKDIR /web

COPY app/package.json app/package-lock.json ./
RUN npm ci

COPY app/ ./
RUN npm run build

FROM nginx:1.27-alpine

COPY --from=builder /web/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
