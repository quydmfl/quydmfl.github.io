resource "cloudflare_record" "app_a" {
  zone_id = var.cloudflare_zone_id
  name    = "app-a"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.homelab.id}.cfargotunnel.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_record" "app_b" {
  zone_id = var.cloudflare_zone_id
  name    = "app-b"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.homelab.id}.cfargotunnel.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
}
