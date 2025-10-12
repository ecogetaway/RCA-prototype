# Render Deployment Guide for RCA Backend

This guide explains how to deploy the RCA backend to Render.com.

## Prerequisites

1. A Render.com account
2. Your code pushed to a Git repository (GitHub, GitLab, or Bitbucket)

## Deployment Steps

### 1. Create a New Web Service

1. Log into your Render dashboard
2. Click "New +" and select "Web Service"
3. Connect your Git repository
4. Select the repository containing this code

### 2. Configure Service Settings

**Basic Settings:**
- **Name**: `rca-backend` (or your preferred name)
- **Environment**: `Docker`
- **Region**: Choose closest to your users
- **Branch**: `main` (or your default branch)

**Build & Deploy:**
- **Dockerfile Path**: `./Dockerfile`
- **Docker Context**: `.` (current directory)

**Health Check:**
- **Health Check Path**: `/health`

### 3. Environment Variables

Set the following environment variables in your Render service:

```
PORT=8080
LISTEN_ADDRESS=0.0.0.0:8080
DATA_DIR=/app/data
DISABLE_USAGE_STATISTICS=true
DO_NOT_CHECK_FOR_DEPLOYMENTS=true
DO_NOT_CHECK_FOR_UPDATES=true
URL_BASE_PATH=/
AUTH_ANONYMOUS_ROLE=admin
AUTH_BOOTSTRAP_ADMIN_PASSWORD=admin
```

### 4. Disk Storage (Optional)

If you need persistent storage:
1. Go to "Disk" tab in your service settings
2. Add a disk with:
   - **Name**: `data`
   - **Mount Path**: `/app/data`
   - **Size**: 1GB (or as needed)

### 5. Deploy

1. Click "Create Web Service"
2. Render will automatically build and deploy your application
3. Monitor the build logs for any issues

## Configuration

The application uses a `config.yaml` file with the following key settings:

- **Port**: Configured to use Render's PORT environment variable
- **Data Directory**: Set to `/app/data` for persistent storage
- **Authentication**: Anonymous access enabled for API testing
- **Cache**: 30-day TTL for performance
- **GRPC**: Disabled for simpler deployment

## API Endpoints

Once deployed, your API will be available at:

- **Health Check**: `GET /health`
- **API Base**: `GET /api/`
- **Authentication**: `POST /api/login`
- **Projects**: `GET /api/project/`

## Troubleshooting

### Common Issues

1. **Build Fails**: Check that all Go dependencies are properly specified in `go.mod`
2. **Port Issues**: Ensure your application listens on the PORT environment variable
3. **Health Check Fails**: Verify the `/health` endpoint is accessible
4. **Database Issues**: The application uses SQLite by default, which is file-based

### Logs

Check the Render service logs for detailed error messages:
1. Go to your service dashboard
2. Click on "Logs" tab
3. Review build and runtime logs

### Environment Variables

If you need to modify configuration:
1. Go to "Environment" tab in your service settings
2. Add or modify environment variables
3. Redeploy the service

## Security Considerations

- Change the default admin password in production
- Configure proper authentication for production use
- Set up proper CORS settings if serving a frontend
- Consider using a managed database for production workloads

## Scaling

For production workloads:
- Consider upgrading to a paid Render plan
- Set up a managed PostgreSQL database
- Configure proper monitoring and alerting
- Implement proper backup strategies
