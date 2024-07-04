# ecr-event-lambda
a lambda in go to forward a ecr push event to a webhook receiver
## This is a lambda that is triggered by an ecr push event
It will notify a listening application server that a new image is available.
For this to work a few things need to be setup first
* First create a policy that only allows the lambda to `secretsmanager:GetSecretValue` for the specific resource (secret) as well as access to logs etc
* Create a secret in secrets manager that contains the secret the application server
* Create a role that allows the lambda to assume the policy created in the first step
* Create a lambda that uses the role created in the previous step
* Create a rule in cloudwatch that triggers the lambda when an ecr push event occurs

`GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go`
