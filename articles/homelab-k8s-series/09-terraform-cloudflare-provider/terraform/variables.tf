variable "cloudflare_zone_id" {
  type        = string
  description = "Zone ID của domain trên Cloudflare"
}

variable "domain" {
  type    = string
  default = "example.com"
}

variable "cloudflare_account_id" {
  type = string
}
