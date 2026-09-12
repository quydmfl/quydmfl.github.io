resource "kubernetes_config_map" "app_config" {
  metadata {
    name      = "app-config"
    namespace = kubernetes_namespace.app.metadata[0].name
  }
  data = {
    APP_ENV  = "homelab"
    LOG_LEVEL = "debug"
  }
}

resource "kubernetes_secret" "app_secret" {
  metadata {
    name      = "app-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }
  data = {
    DB_PASSWORD = "sup3rS3cret!"
  }
}
