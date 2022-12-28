FROM node:18 as build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:1.23

COPY nginx.conf /etc/nginx/user.conf.d/default.conf

COPY --from=build /app/dist/ /usr/share/nginx/html
