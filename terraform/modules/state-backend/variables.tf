variable "state_bucket_name" {
  description = "Name of the S3 bucket for Terraform state storage"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*[a-z0-9]$", var.state_bucket_name))
    error_message = "Bucket name must start and end with lowercase letter or number, and contain only lowercase letters, numbers, and hyphens."
  }
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
