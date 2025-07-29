# go-serverless-framework

## Project Description: 
This project provides an organizational framework for rapidly deploying AWS Serverless Full-Stack applications. It's designed to streamline the development and deployment of serverless solutions, integrating seamlessly with commonly used AWS services.

### Key Features:
Example API: Includes a practical example API within the api folder, demonstrating typical lambda integrations and event handlers.

### Comprehensive AWS CDK Integration:
The cdk folder contains all the necessary code to deploy serverless applications, featuring integrations with services such as:

- CloudFront

- API Gateway

- DynamoDB

- S3

- SQS

- SES

- Cognito

### Multi-Stage Environment Support: 
Supports various staging environments including Local, Dev, and Production, facilitating a robust development workflow.

### Customization & Flexibility:
This repository is not intended as an out-of-the-box solution, but rather as a highly customizable foundation. Developers are encouraged to adapt and extend this framework to meet their specific project requirements and architectural needs.

## Important Notice
**This framework is currently under active development and testing. As such, it may contain undiscovered security vulnerabilities and bugs.**

Your contributions are highly valued! If you identify any vulnerabilities, bugs, or opportunities for optimization, please feel free to create a Pull Request or create a bug report.

## Missing Frontend Assets
Please note that, althought their cdk IaC is present in the cdk code base, their  folders ("frontend" and "admin") are not included in this repository. You'll need to provide these assets and integrate them according to your specific project requirements.

## Tech stack
This project is designed using the following tech stack:
- node v22-lts
- AWS CDK (language: golang)
- Go v1.23.1

## Technical description
This cdk implementation consists on 3 stacks self containing all necessary constucts to build the required infrastructure for the project, being:
- go-serverless-framework-backend
- go-serverless-framework-frontend
- go-serverless-framework-admin

## High Level Architecture
![alt text](/system-design/hl_arch.png)

## Run the project locally with Localstack
- Populate cdk/constants/constants.go with values adapted to your own project
- Install localstack (https://docs.localstack.cloud/aws/getting-started/installation/)
- Add new localstack profile to your local aws credentials and config (~/.aws/credentials, ~/.aws/config). You config file should like like this:
```shell
    [default]
    region = xxxxxx
    output = json
    [profile localstack]
    region = xxxxxx
    output = json
```
and, credentials file should look like this:
```shell
    [default]
    aws_access_key_id = XXXYYYZZZXXXYYYZZZ
    aws_secret_access_key = xxxyyyzzz1111xxxyyyzzz111
    [localstack]
    aws_access_key_id = test
    aws_secret_access_key = test
```
- execute the following commands to build the backend/api:
1. `localstack start` 
2. `cdklocal bootstrap --context buildFrontend=false --context buildAdmin=false --context buildUserPortal=false --profile=localstack aws://000000000000/us-east-1`
3. Create an hosted zone in your local localstack environment (this is an explicit dependency of this project) 
`awslocal route53 create-hosted-zone --name "upwigo.com" --caller-reference "local-dev-$(date +%s)"`
4. `cdklocal synth --context buildFrontend=false --context buildAdmin=false --context buildUserPortal=false "Local/go-serverless-framework-backend" --profile=localstack`
5. `cdklocal deploy --context buildFrontend=false --context buildAdmin=false --context buildUserPortal=false "Local/go-serverless-framework-backend" --profile=localstack`

## Deploy to AWS
- Create an hosted zone in your AWS account
- This project supports the usage of different stages of deployment: local, dev and production. You should populate the file cdk/constants/constants.go with values adapted to your own project

### Deploy command Dev
npx cdk deploy --context buildFrontend=false --context buildAdmin=false --context buildUserPortal=false "Dev/*" --profile=your-dev-profile

### Deploy command Prod
npx cdk deploy --context buildFrontend=false --context buildAdmin=false --context buildUserPortal=false "Prod/*" --profile=your-prod-profile

## After deploying
- Create the admin and user users in Cognito (for testing, tick email verified)
- You will receive an email from SES to the email you have set in  cdk/constants/constants.go - `applicationEmail`, prompting you to verify you email address - click on the link to verify it
- you are all set :)

# ONGOING WORK
## CI/CD
- create CICD using GitHub Actions:
- on **ready to merge**: 
    -  build
    - test
    - lint
    - automate change log based on conventional commits
    - synth -> insert synth out in PR description
- on **merge PR**
    - deploy
## Monitoring
- centralize lambda logs into a single log group
- create alarms and triggers for abnormally large number of lambda invocations
- create alarms and triggers for billing budgets and price anomaly detection
    - teardown
- add email trigger upon >= 500 errors
- create individualized lambda latency charts and add to cw dashboard 

## Unit Tests
- Add missing unit tests in repository package
- Add end to end unit tests for lambda integrations

## Documentation
- Add automatic api documentation (swagger?...)
- Update Access Patterns to faithfuly mirror the project current state

## Code
- Use context to handle abnormal latency of external API requests
- Use BatchSendMessage when using SQS SDK when appropriate
- Change deeper scoped function variables names to be shorter
- Change interface names to "prefix-er"
- uniformize and standardize "enums"
