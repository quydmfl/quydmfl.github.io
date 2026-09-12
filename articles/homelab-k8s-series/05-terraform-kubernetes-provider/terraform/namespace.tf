resource "kubernetes_namespace" "app" {
  metadata {
    name = "tf-demo"
    labels = {
      managed-by = "terraform"
    }
  }
}
