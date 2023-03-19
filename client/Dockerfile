# stage 1
FROM node:14.18.0-alpine as build

WORKDIR /app

# Copy vue app source code
COPY . /app/

# Install yarn and build project
RUN yarn cache clean
RUN yarn install
RUN echo "VITE_API_ENDPOINT=$VITE_API_ENDPOINT" >> /app/.env
RUN yarn run build


# stage 2
FROM nginx
RUN mkdir /app
COPY --from=build /app/dist /app
COPY nginx.conf /etc/nginx/nginx.conf