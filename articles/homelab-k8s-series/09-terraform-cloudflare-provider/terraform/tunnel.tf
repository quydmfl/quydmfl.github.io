resource "random_id" "tunnel_secret" {
  byte_length = 35
}

resource "cloudflare_zero_trust_tunnel_cloudflared" "homelab" {
  account_id = var.cloudflare_account_id
  name       = "homelab"
  secret     = random_id.tunnel_secret.b64_std
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "homelab" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.homelab.id

  config {
    ingress_rule {
      hostname = "app-a.${var.domain}"
      service  = "http://app-a.routing-demo.svc.cluster.local"
    }
    ingress_rule {
      hostname = "app-b.${var.domain}"
      service  = "http://app-b.routing-demo.svc.cluster.local:5678"
    }
    ingress_rule {
      service = "http_status:404"
    }
  }
}
