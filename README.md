# go-serverless-framework

## Tech stack
This project is designed using the following tech stack:
- node v22-lts
- Angular v19
- TailwindCSS
- Astro framework
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

# Access Patterns
# DynamoDB Access Pattern Documentation

**Table Name:** `MainTable`

**Last Updated:** `2025-04-11`

## 1. Introduction

This document outlines the access patterns for the `MainTable` DynamoDB table.


## 2. Access Patterns

### 2.1. Access Pattern 1

* **Name:** Get Applications Between Dates
* **Description:** Fetches Applications between date A and date B
* **Frequency:** Medium
* **Query Type:**
    * `Query`
* **Key Conditions (if applicable):**
    * **Partition Key Condition:**  `GSI_PK = APP#`
    * **Sort Key Condition:** (e.g., `SK = BETWEEN DATE_A AND DATE_B`)
    * **Index Used (if applicable):** MainTableGSI
* **Expected Throughput (Read/Write Capacity Units - RCUs/WCUs):** (10 * 200bytes)/(4KB*2) = 0.5 RCU
* **Note:** Same pattern is used for C Applications

### 2.1. Access Pattern 2

* **Description:** Get document with type {type} by application {id}
* **Frequency:** Medium
* **Query Type:** `GetItem`
* **Key Conditions:**
    * **Partition Key Condition:** `PK = ApplicationID`
    * **Sort Key Condition:** `SK = DocumentType`
* **Index Used:** `MainTable`
* **Expected Throughput:**: 0.5 RCU

### 2.2. Access Pattern 3

* **Description:** Get posts by language
* **Frequency:** High
* **Query Type:** `Query`
* **Key Conditions:**
    * **Partition Key Condition:** `GSI_PK = P#<lang>`
* **Index Used:** `MainTableGSI`
* **Expected Throughput:**: 0.5 RCU

### 2.3. Access Pattern 4

* **Description:** Get Application ID by Email
* **Frequency:** Low
* **Query Type:** `Query`
* **Key Conditions:**
    * **Partition Key Condition:** `Email = <email>`
* **Index Used:** `MainTableGSI_2`
* **Expected Throughput:**: 0.5 RCU

### 2.3. Access Pattern 5

* **Description:** Get Documents By Application ID
* **Frequency:** Low
* **Query Type:** `Query`
* **Key Conditions:**
    * **Partition Key Condition:** `PK = ApplicationID`
    * **Sort Key Condition:** `SK = STARTS_WITH "D`
* **Index Used:** `MainTable`
* **Expected Throughput:**: 0.5 RCU

### 2.4. Access Pattern 6

* **Description:** Get Messages by Application ID
* **Frequency:** Low
* **Query Type:** `Query`
* **Key Conditions:**
    * **Partition Key Condition:** `PK = ApplicationID`
     * **Sort Key Condition:** (e.g., ``)
* **Index Used:** `MainTable`
* **Expected Throughput:**: 1 RCU

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