#!/bin/bash

# AWS Infrastructure Deployment Script
# This script deploys the scheduled items API to AWS using CDK

set -e

echo "🚀 Starting AWS infrastructure deployment..."

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI not configured. Please run 'aws configure' first."
    exit 1
fi

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
    echo "❌ AWS CDK not found. Please install with: npm install -g aws-cdk"
    exit 1
fi

# Get the script directory and navigate to CDK directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CDK_DIR="$SCRIPT_DIR/cdk"

echo "📁 CDK directory: $CDK_DIR"
cd "$CDK_DIR"

# Clean any existing cdk.out directories
echo "🧹 Cleaning existing CDK output..."
rm -rf cdk.out .cdk.staging

echo "📦 Installing CDK dependencies..."
go mod tidy

echo "🔧 Bootstrapping CDK (if needed)..."
cdk bootstrap

echo "🔨 Building and deploying infrastructure..."
cdk deploy --require-approval never

echo "✅ Deployment complete!"
echo ""
echo "📋 Next steps:"
echo "1. Note the LoadBalancerDNS output from the deployment"
echo "2. Test your API using: curl http://<LoadBalancerDNS>/scheduled-items"
echo "3. The RDS database will automatically scale based on usage"
echo ""
echo "💰 Cost optimization tips:"
echo "- RDS Serverless v2 scales down to 0.5 ACU when idle"
echo "- Single NAT Gateway reduces costs vs multi-AZ setup"  
echo "- Consider setting up CloudWatch alarms for cost monitoring"