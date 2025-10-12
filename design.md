# RCA Prototype Design

## Overview
Build a web application for Managed Service Providers (MSPs) and IT Infrastructure service providers, leveraging coroot for automated root cause analysis (RCA) with AI-driven diagnostics, integrated with Keep's alert management and AWS Agentic AI technologies.

## Architecture
- **Backend**: Fork of coroot (Go), extended with AI diagnostics module and Keep integration.
- **Frontend**: Vue.js (based on coroot's frontend), customized for MSP workflows.
- **Storage**: ClickHouse for metrics/logs, PostgreSQL for application data.
- **AI Module**: AWS Bedrock AgentCore for agent orchestration, Strands Agents SDK for autonomous agents, integrated with Keep's AI backends.

## Components
- **Coroot Core**: Handles data collection, basic RCA, dashboards.
- **Agentic AI Layer**: Bedrock AgentCore for dynamic alert correlation and decision-making; Strands Agents for infrastructure drift detection and summarization.
- **Keep Integration**: Connects alerts and workflows with Keep's platform.
- **MSP Frontend**: Custom UI for asset management, client views, reports.

## Features
- Asset monitoring dashboard with agent-driven insights.
- Real-time malfunction alerts via AgentCore orchestration.
- AI-enhanced RCA with predictive insights from Strands Agents.
- Multi-tenant support for MSP clients.

## Implementation Plan
1. Fork coroot and integrate with Keep.
2. Deploy Bedrock AgentCore for alert processing.
3. Implement Strands Agents for autonomous workflows.
4. Develop custom frontend components.
5. Add MSP-specific features like client isolation.