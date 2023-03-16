# stage 1
FROM node:14.18.0-alpine as build

WORKDIR /app

# Copy vue app source code
COPY . /app/

# Install yarn and build project
RUN yarn cache clean
RUN yarn install
RUN yarn run build

# stage 2
FROM nginx:1.19-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080