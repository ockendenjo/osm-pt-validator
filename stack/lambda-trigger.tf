module "lambda_trigger" {
  source                   = "github.com/ockendenjo/tfmods//lambda"
  aws_env                  = var.env
  name                     = "trigger"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "osmptv"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["trigger"]
  alarm_topic_arn          = aws_sns_topic.alarms.arn

  environment = {
    S3_BUCKET_NAME = aws_s3_bucket.data.id
    QUEUE_URL      = module.sqs_validate_rm_events.queue_url
  }
}

resource "aws_cloudwatch_event_rule" "daily_schedule" {
  name                = "osmptv-${var.env}-daily-schedule"
  description         = "Trigger validation once per day at 23:05 UTC"
  schedule_expression = "cron(5 23 * * ? *)"
}

resource "aws_cloudwatch_event_target" "trigger_lambda" {
  rule      = aws_cloudwatch_event_rule.daily_schedule.name
  target_id = "TriggerLambda"
  arn       = module.lambda_trigger.arn
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_trigger.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily_schedule.arn
}
