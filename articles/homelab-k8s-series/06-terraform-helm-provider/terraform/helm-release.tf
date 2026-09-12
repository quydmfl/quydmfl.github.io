resource "helm_release" "webapp" {
  name      = "myapp"
  namespace = kubernetes_namespace.app.metadata[0].name
  chart     = "./webapp"

  values = [
    file("./webapp/values-dev.yaml")
  ]

  depends_on = [kubernetes_namespace.app]
}
