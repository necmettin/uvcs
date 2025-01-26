FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Expose Vite dev server port
EXPOSE 5173

# Start development server with host set to allow external access
CMD ["npm", "run", "dev", "--", "--host"] 