# AWS Infrastructure Deployment

This directory contains AWS CDK infrastructure code to deploy your Go REST API to AWS using:

- **AWS Fargate** - Serverless container hosting
- **Amazon RDS Serverless v2** - Auto-scaling PostgreSQL database
- **Application Load Balancer** - HTTP traffic distribution
- **VPC** - Isolated network environment

## Prerequisites

1. **AWS CLI** configured with appropriate credentials:
   ```bash
   aws configure
   ```

2. **AWS CDK** installed globally:
   ```bash
   npm install -g aws-cdk
   ```

3. **Go** 1.21+ installed for CDK

## Deployment

### Quick Deploy
```bash
./deploy.sh
```

### Manual Deploy
```bash
cd infrastructure/cdk
go mod tidy
cdk bootstrap
cdk deploy
```

## Architecture Overview

```
Internet → ALB → Fargate → RDS Serverless v2
           │
           └── CloudWatch Logs
```

### Components

- **VPC**: 2 AZs with public, private, and database subnets
- **ALB**: Internet-facing load balancer on port 80
- **Fargate**: Container service running your Go API
- **RDS**: Aurora PostgreSQL Serverless v2 (0.5-1 ACU)
- **Secrets Manager**: Database credentials

## Cost Optimization Features

- **RDS Serverless v2**: Scales down to 0.5 ACU when idle
- **Single NAT Gateway**: Reduces costs vs multi-AZ setup
- **Fargate**: Pay only for running containers
- **CloudWatch Logs**: 1-week retention to limit costs

## Environment Variables

The CDK automatically configures these environment variables for your container:

- `USE_POSTGRES_DB=true`
- `DB_HOST` (from RDS endpoint)
- `DB_PORT` (from RDS)
- `DB_NAME` (from RDS)
- `DB_USER` (from Secrets Manager)
- `DB_PASSWORD` (from Secrets Manager)

## Monitoring

- CloudWatch Logs: `/aws/ecs/scheduled-items`
- RDS Metrics: Available in CloudWatch
- ALB Metrics: Request counts, latency, errors

## Clean Up

To avoid ongoing costs:
```bash
cd infrastructure/cdk
cdk destroy
```

## Troubleshooting

### Container Won't Start
- Check CloudWatch logs: `/aws/ecs/scheduled-items`
- Verify database connectivity
- Check environment variables

### Database Connection Issues
- Ensure security groups allow traffic
- Verify RDS is in available state
- Check Secrets Manager for credentials

### High Costs
- Monitor RDS ACU usage in CloudWatch
- Consider reducing Fargate CPU/memory
- Review NAT Gateway data transfer charges