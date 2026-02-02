terraform {
  backend "s3" {
    # Pass config using -backend-config via CLI
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.0"
    }
  }
}

provider "aws" {

  default_tags {
    tags = {
      Project     = "osmptv"
      Environment = "pro"
    }
  }
}
