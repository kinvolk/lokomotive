output "clc_config" {
  value = data.ct_config.config.rendered
}

output "bootstrap_token" {
  value = local.bootstrap_token
}
