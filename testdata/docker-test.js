// JavaScript file that uses Docker environment variables
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const port = process.env.PORT;

// This variable is declared in Dockerfile but not in .env
const nodeEnv = process.env.NODE_ENV;

// This variable is in .env but not in Dockerfile
const secretKey = process.env.SECRET_KEY;

console.log(`API Key: ${apiKey}`);
console.log(`Database: ${dbUrl}`);
console.log(`Port: ${port}`);
console.log(`Environment: ${nodeEnv}`);
console.log(`Secret: ${secretKey}`);
