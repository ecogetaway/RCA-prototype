# Modifications to Coroot

## Backend Extensions
- Integrate with Keep for alert management.
- Add Bedrock AgentCore for dynamic agent orchestration in alert processing.
- Extend api/rca.go to include agent-driven insights in RCA response.
- Implement Strands Agents for autonomous workflows (e.g., drift detection).

## Frontend Customizations
- Modify front/src/ to add MSP-specific views.
- Add client management components.
- Enhance RCA view with agent suggestions and Keep integrations.

## AI Integration
- Use AWS Bedrock for foundation models.
- Strands Agents SDK for building model-driven agents.
- AgentCore Gateway for transforming data sources into agent tools.
- Integrate with Keep's existing AI backends.

## Workflow Changes
- Custom RCA workflows for MSP scenarios with agent-based decision-making.
- Add project-level client isolation.
- Multi-step alert processing via autonomous agents.