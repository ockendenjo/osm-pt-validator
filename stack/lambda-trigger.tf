module "lambda_trigger" {
  source                   = "github.com/ockendenjo/tfmods//lambda"
  aws_env                  = var.env
  name                     = "trigger"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "osmptv"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["trigger"]

  environment = {
    S3_BUCKET_NAME = aws_s3_bucket.data.id
    QUEUE_URL      = module.sqs_validate_rm_events.queue_url
  }
}
