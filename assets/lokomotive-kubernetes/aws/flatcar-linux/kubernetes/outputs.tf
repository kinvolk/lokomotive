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

# Outputs for custom load balancing

output "nlb_id" {
  description = "ARN of the Network Load Balancer"
  value       = aws_lb.nlb.id
}

output "worker_target_group_http" {
  description = "ARN of a target group of workers for HTTP traffic"
  value       = module.workers.target_group_http
}

output "worker_target_group_https" {
  description = "ARN of a target group of workers for HTTPS traffic"
  value       = module.workers.target_group_https
}

# values.yaml content for all deployed charts.
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

output "flannel_values" {
  value = module.bootkube.flannel_values
}

output "kube-router_values" {
  value = module.bootkube.kube-router_values
}
