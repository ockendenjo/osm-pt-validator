module "lambda_split_relation" {
  source                   = "github.com/ockendenjo/tfmods//lambda"
  aws_env                  = var.env
  name                     = "validate-rm"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "osmptv"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["validate-rm"]
  alarm_topic_arn          = aws_sns_topic.alarms.arn

  environment = {
    QUEUE_URL  = module.sqs_validate_route_events.queue_url
    TOPIC_ARN  = aws_sns_topic.invalid_relations.arn
    USER_AGENT = "https://github.com/ockendenjo/osm-pt-validator"
  }
}

module "iam_sns_lambda_split_relation" {
  source  = "github.com/ockendenjo/tfmods//iam-sns"
  role_id = module.lambda_split_relation.role_id
  sns_arns = [
    aws_sns_topic.invalid_relations.arn,
  ]
}

module "iam_sqs_lambda_split_relation" {
  source  = "github.com/ockendenjo/tfmods//iam-sqs"
  role_id = module.lambda_split_relation.role_id
  sqs_arns = [
    module.sqs_validate_route_events.queue_arn,
  ]
}

module "sqs_eventsource_lambda_split_relation" {
  source = "github.com/ockendenjo/tfmods//lambda-sqs-source"
  lambda = module.lambda_split_relation
  queue  = module.sqs_validate_rm_events
}
