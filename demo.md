# RCA Prototype Demo

This prototype demonstrates an AI-enhanced Root Cause Analysis (RCA) system for Managed Service Providers (MSPs), integrating coroot with Keep alerts, AWS Bedrock AgentCore, and Strands Agents.

## Features Highlighted

1. **Keep Integration**: Alerts are sent to Keep for centralized alert management.
2. **AWS Bedrock AgentCore**: Orchestrates alert processing with AI agents.
3. **Strands Agents**: Provides autonomous workflows for drift detection and summarization.
4. **AI-Enhanced RCA**: RCA views include agent-driven insights.
5. **MSP Frontend**: Custom UI for client management and multi-tenant support.
6. **Multi-Tenant Support**: Projects are isolated by client_id.

## Setup Instructions

### Prerequisites
- Go 1.23+
- Node.js for frontend
- Docker for deployment (optional)
- AWS credentials for Bedrock (optional, mocked if not configured)
- Keep server running (optional, mocked if not configured)

### Backend Setup
1. Clone the repository: `git clone https://github.com/coroot/coroot.git`
2. Navigate to coroot: `cd coroot`
3. Install dependencies: `go mod tidy`
4. Build: `go build`
5. (Optional) Configure config.yaml with Keep URL, AWS settings; if not configured, mocking is used.

### Frontend Setup
1. Navigate to front: `cd front`
2. Install dependencies: `npm install`
3. Build: `npm run build`

### Deployment
#### Local
1. Build Docker image: `docker build -t rca-prototype .`
2. Run: `docker run -p 8080:8080 rca-prototype`

#### Production
- **Frontend (Netlify)**: Deploy `coroot/front/dist` directory to Netlify. Set environment variable `VUE_APP_API_BASE_URL` to Render backend URL.
- **Backend (Render)**: Deploy Go app from `coroot/` directory to Render. Use build command `go build -o rca-backend`. Set start command `./rca-backend`. Configure environment variables like `KEEP_URL`, `BEDROCK_AGENT_ID`, etc. (leave empty for mock).

## Demo Steps

1. **Start the Application**: Run the backend server.
2. **Access Frontend**: Open http://localhost:8080
3. **Create Project**: Add a project with client_id for multi-tenancy.
4. **Configure Integrations**: Set up Prometheus, Keep, AWS Bedrock.
5. **Simulate Incident**: Trigger an alert in the system.
6. **View RCA**: Navigate to RCA view for an application.
   - See AI Agent Insights section with Strands summarization.
7. **Check Logs**: Verify mock alerts logged (if Keep not configured).
8. **Client Management**: Access Client Management view for MSP features.

## Key Files Modified
- `keep/keep.go`: Keep client for alerts.
- `bedrock/bedrock.go`: AWS Bedrock integration.
- `strands/strands.go`: Strands agents for workflows.
- `api/rca.go`: Extended with agent insights.
- `views/RCA.vue`: Added AI insights display.
- `views/ClientManagement.vue`: MSP client management.
- `db/project.go`: Added client_id for multi-tenancy.

## Validation Checklist
- [ ] Backend compiles without errors.
- [ ] Frontend builds successfully.
- [ ] Docker image builds.
- [ ] Application starts and serves UI.
- [ ] RCA shows agent insights.
- [ ] Mock alerts logged (if Keep not configured).
- [ ] Multi-tenant isolation works.