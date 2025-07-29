## Get access token programatically
```shell
aws cognito-idp admin-initiate-auth \
  --user-pool-id xxxxx \
  --client-id xxxxxx \
  --auth-flow ADMIN_USER_PASSWORD_AUTH \
  --auth-parameters USERNAME=example@email.com,PASSWORD=test123# \
  --region eu-central-1
```
