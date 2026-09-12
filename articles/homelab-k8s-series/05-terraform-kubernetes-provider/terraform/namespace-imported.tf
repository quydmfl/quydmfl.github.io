resource "kubernetes_namespace" "imported" {
  metadata {
    name = "tf-imported"
  }
}
