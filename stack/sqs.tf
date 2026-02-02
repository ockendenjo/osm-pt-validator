module "sqs_validate_route_events" {
  source       = "github.com/ockendenjo/tfmods//sqs"
  aws_env      = var.env
  name         = "validate-route-events"
  project_name = "osmptv"
}

module "sqs_validate_rm_events" {
  source       = "github.com/ockendenjo/tfmods//sqs"
  aws_env      = var.env
  name         = "validate-rm-events"
  project_name = "osmptv"
}
