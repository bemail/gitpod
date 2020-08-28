resource "null_resource" "kubeconfig" {
  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --name $CLUSTER"
    environment = {
      CLUSTER = var.kubernetes.cluster_name
    }
  }
  depends_on = [
    module.kubernetes
  ]
}
