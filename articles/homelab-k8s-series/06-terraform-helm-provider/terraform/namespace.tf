resource "kubernetes_namespace" "app" {
  metadata {
    name = "tf-helm-demo"
  }
}
