## Get access token programatically
```shell
aws cognito-idp admin-initiate-auth \
  --user-pool-id eu-central-1_tbsdJnhqN \
  --client-id 52iufiq99be4dt42glaklvu105 \
  --auth-flow ADMIN_USER_PASSWORD_AUTH \
  --auth-parameters USERNAME=andre@upwigo.com,PASSWORD=test123# \
  --region eu-central-1
```