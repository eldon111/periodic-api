package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InfrastructureStackProps struct {
	awscdk.StackProps
}

func NewInfrastructureStack(scope constructs.Construct, id string, props *InfrastructureStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC
	vpc := awsec2.NewVpc(stack, jsii.String("PeriodicApiVPC"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(1), // Cost optimization - single NAT gateway
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Database"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Create RDS Serverless v2 PostgreSQL
	dbCluster := awsrds.NewDatabaseCluster(stack, jsii.String("PeriodicApiDB"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_15_4(),
		}),
		ServerlessV2MaxCapacity: jsii.Number(1),   // Auto-scale up to 1 ACU
		ServerlessV2MinCapacity: jsii.Number(0.5), // Scale down to 0.5 ACU
		Writer: awsrds.ClusterInstance_ServerlessV2(jsii.String("writer"), &awsrds.ServerlessV2ClusterInstanceProps{
			PubliclyAccessible: jsii.Bool(false),
		}),
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		DefaultDatabaseName: jsii.String("periodic_api_db"),
		DeletionProtection:  jsii.Bool(false),             // Set to true for production
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY, // Change for production
	})

	// Create ECS Cluster
	cluster := awsecs.NewCluster(stack, jsii.String("PeriodicApiCluster"), &awsecs.ClusterProps{
		Vpc:                 vpc,
		ContainerInsightsV2: awsecs.ContainerInsights_ENABLED,
	})

	// Create Task Definition
	taskDefinition := awsecs.NewFargateTaskDefinition(stack, jsii.String("PeriodicApiTaskDef"), &awsecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(512),
		Cpu:            jsii.Number(256),
	})

	// Add container to task definition
	container := taskDefinition.AddContainer(jsii.String("PeriodicApiContainer"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromAsset(jsii.String("../../"), &awsecs.AssetImageProps{
			File:      jsii.String("Dockerfile"),
			BuildArgs: &map[string]*string{},
		}),
		Environment: &map[string]*string{
			"USE_POSTGRES_DB": jsii.String("true"),
		},
		Secrets: &map[string]awsecs.Secret{
			"DB_HOST":     awsecs.Secret_FromSecretsManager(dbCluster.Secret(), jsii.String("host")),
			"DB_PORT":     awsecs.Secret_FromSecretsManager(dbCluster.Secret(), jsii.String("port")),
			"DB_NAME":     awsecs.Secret_FromSecretsManager(dbCluster.Secret(), jsii.String("dbname")),
			"DB_USER":     awsecs.Secret_FromSecretsManager(dbCluster.Secret(), jsii.String("username")),
			"DB_PASSWORD": awsecs.Secret_FromSecretsManager(dbCluster.Secret(), jsii.String("password")),
		},
		Logging: awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("periodic-api"),
			LogRetention: awslogs.RetentionDays_ONE_WEEK,
		}),
	})

	container.AddPortMappings(&awsecs.PortMapping{
		ContainerPort: jsii.Number(8080),
		Protocol:      awsecs.Protocol_TCP,
	})

	// Create Fargate Service
	service := awsecs.NewFargateService(stack, jsii.String("PeriodicApiService"), &awsecs.FargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: taskDefinition,
		DesiredCount:   jsii.Number(1),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		EnableExecuteCommand: jsii.Bool(true), // For debugging
	})

	// Create Application Load Balancer
	alb := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, jsii.String("PeriodicApiALB"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
	})

	// Create Target Group
	targetGroup := awselasticloadbalancingv2.NewApplicationTargetGroup(stack, jsii.String("PeriodicApiTargetGroup"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		Port:       jsii.Number(8080),
		Protocol:   awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		Vpc:        vpc,
		TargetType: awselasticloadbalancingv2.TargetType_IP,
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Path:                    jsii.String("/scheduled-items"),
			HealthyHttpCodes:        jsii.String("200"),
			Interval:                awscdk.Duration_Seconds(jsii.Number(30)),
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(5)),
			HealthyThresholdCount:   jsii.Number(2),
			UnhealthyThresholdCount: jsii.Number(3),
		},
	})

	// Add service to target group
	targetGroup.AddTarget(service)

	// Add listener to ALB
	alb.AddListener(jsii.String("PeriodicApiListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
		Port:                jsii.Number(80),
		Protocol:            awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		DefaultTargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{targetGroup},
	})

	// Allow service to connect to database
	dbCluster.Connections().AllowDefaultPortFrom(service, jsii.String("Allow ECS to connect to RDS"))

	// Output the ALB DNS name
	awscdk.NewCfnOutput(stack, jsii.String("LoadBalancerDNS"), &awscdk.CfnOutputProps{
		Value:       alb.LoadBalancerDnsName(),
		Description: jsii.String("DNS name of the load balancer"),
	})

	// Output the database endpoint
	awscdk.NewCfnOutput(stack, jsii.String("DatabaseEndpoint"), &awscdk.CfnOutputProps{
		Value:       dbCluster.ClusterEndpoint().Hostname(),
		Description: jsii.String("RDS cluster endpoint"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfrastructureStack(app, "PeriodicApiInfrastructureStack", &InfrastructureStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
