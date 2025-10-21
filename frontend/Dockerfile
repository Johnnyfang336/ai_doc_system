# Use official Node.js image as build environment
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package.json
COPY package.json ./

# Copy source code first
COPY . .

# Install dependencies and build
RUN npm install && npm run build

# Use nginx as production environment
FROM nginx:alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy build results from build stage
COPY --from=builder /app/build /usr/share/nginx/html

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]