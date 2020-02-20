provider "matchbox" {
  ca          = "${file(pathexpand("~/pxe-testbed/.matchbox/ca.crt"))}"
  client_cert = "${file(pathexpand("~/pxe-testbed/.matchbox/client.crt"))}"
  client_key  = "${file(pathexpand("~/pxe-testbed/.matchbox/client.key"))}"
  endpoint    = "matchbox.example.com:8081"
}
