resource "aws_sns_topic" "invalid_relations" {
  name = "osmptv-${var.env}-invalid-relations"
}

resource "aws_sns_topic_subscription" "invalid_relations" {
  for_each  = toset(var.invalid_relation_emails)
  endpoint  = each.key
  protocol  = "email"
  topic_arn = aws_sns_topic.invalid_relations.arn
}

resource "aws_sns_topic" "alarms" {
  name = "osmptv-${var.env}-alarms"
}

resource "aws_sns_topic_subscription" "alarms" {
  for_each  = toset(var.alarm_emails)
  endpoint  = each.key
  protocol  = "email"
  topic_arn = aws_sns_topic.alarms.arn
}
