{
  "id": "${id}",
  "metadata": {
    "facility": {
      "facility_code": "${facility_code}",
      "plan_slug": "${plan_slug}",
      "plan_version_slug": ""
    },
    "instance": {},
    "state": ""
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "ip": {
            "address": "${address}",
            "gateway": "${gateway}",
            "netmask": "${netmask}"
          },
          "mac": "${mac}"
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
