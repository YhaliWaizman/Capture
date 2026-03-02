# Simple Dockerfile with ENV and ARG
FROM node:18-alpine

# Build arguments
ARG NODE_ENV=production
ARG BUILD_VERSION

# Environment variables
ENV API_KEY=default_key
ENV DATABASE_URL=postgres://localhost:5432/db
ENV PORT 3000

# Using variables
RUN echo "Building version $BUILD_VERSION"
RUN echo "API Key: ${API_KEY}"

# This variable is used but not declared
RUN echo "Missing: $MISSING_VAR"

WORKDIR /app
COPY . .

CMD ["node", "server.js"]
