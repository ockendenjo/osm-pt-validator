variable "env" {
  description = "Environment name (dev or pro)"
  type        = string

  validation {
    condition     = contains(["dev", "pro"], var.env)
    error_message = "Environment must be either 'dev' or 'pro'."
  }
}

variable "permissions_boundary_arn" {
  type = string
}

variable "lambda_binaries_bucket" {
  type = string
}
