output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

# Outputs for Kubernetes Ingress

output "ingress_dns_name" {
  value       = aws_lb.nlb.dns_name
  description = "DNS name of the network load balancer for distributing traffic to Ingress controllers"
}

output "ingress_zone_id" {
  value       = aws_lb.nlb.zone_id
  description = "Route53 zone id of the network load balancer DNS name that can be used in Route53 alias records"
}

# Outputs for worker pools

output "vpc_id" {
  value       = aws_vpc.network.id
  description = "ID of the VPC for creating worker instances"
}

output "subnet_ids" {
  value       = [aws_subnet.public.*.id]
  description = "List of subnet IDs for creating worker instances"
}

output "worker_security_groups" {
  value       = [aws_security_group.worker.id]
  description = "List of worker security group IDs"
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "ca_cert" {
  value = module.bootkube.ca_cert
}

output "apiserver" {
  value = local.api_server
}

# Outputs for custom load balancing

output "nlb_arn" {
  description = "ARN of the Network Load Balancer"
  value       = aws_lb.nlb.arn
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.bootkube.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value = module.bootkube.kube-apiserver_values
}

output "kubernetes_values" {
  value = module.bootkube.kubernetes_values
}

output "kubelet_values" {
  value = module.bootkube.kubelet_values
}

output "calico_values" {
  value = module.bootkube.calico_values
}

output "lokomotive_values" {
  value = module.bootkube.lokomotive_values
}

output "bootstrap-secrets_values" {
  value = module.bootkube.bootstrap-secrets_values
}

output "node-local-dns_values" {
  value = module.bootkube.node-local-dns_values
}
