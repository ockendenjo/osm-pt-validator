module "lambda_validate_route" {
  source                   = "github.com/ockendenjo/tfmods//lambda"
  aws_env                  = var.env
  name                     = "validate-route"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "osmptv"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["validate-route"]
  alarm_topic_arn          = aws_sns_topic.alarms.arn

  environment = {
    TOPIC_ARN  = aws_sns_topic.invalid_relations.arn
    USER_AGENT = "https://github.com/ockendenjo/osm-pt-validator"
  }
}

module "iam_sns_lambda_validate_route" {
  source  = "github.com/ockendenjo/tfmods//iam-sns"
  role_id = module.lambda_validate_route.role_id
  sns_arns = [
    aws_sns_topic.invalid_relations.arn,
  ]
}

module "sqs_eventsource_validate_route" {
  source = "github.com/ockendenjo/tfmods//lambda-sqs-source"
  lambda = module.lambda_validate_route
  queue  = module.sqs_validate_route_events
}
