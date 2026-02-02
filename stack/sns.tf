resource "aws_sns_topic" "invalid_relations" {
  name = "osmptv-${var.env}-invalid-relations"
}

resource "aws_sns_topic" "alarms" {
  name = "osmptv-${var.env}-alarms"
}
