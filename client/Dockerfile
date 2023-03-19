# stage 1
FROM node:18.15.0-alpine as build

WORKDIR /app

# Copy vue app source code
COPY . /app/

# Install yarn and build project
RUN yarn cache clean
RUN yarn install

# RUN export VITE_API_ENDPOINT="http://localhost:3000/v1"
CMD echo "VITE_API_ENDPOINT=$VITE_API_ENDPOINT" >> /app/.env
RUN yarn run build
# RUN chmod +x /app/start.sh
# ENTRYPOINT ["sh", "/app/start.sh"]

# stage 2
FROM nginx
RUN mkdir /app
COPY --from=build /app/dist /app
COPY nginx.conf /etc/nginx/nginx.conf