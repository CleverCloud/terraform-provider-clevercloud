Expose a Clever Cloud application through an [Otoroshi](https://www.clever.cloud/developers/doc/addons/otoroshi/) add-on.

The route binds a public domain, optionally restricted to a path prefix, to an application reachable inside the Otoroshi networkgroup. The application does not need any public virtual host of its own: traffic goes through Otoroshi.

Both ends are derived rather than hand-written:

* the **backend** is the networkgroup FQDN of the application, built from `application_id` and the add-on networkgroup
* the **frontend** defaults to the `cleverapps.io` domain of the Otoroshi entrypoint

The Otoroshi add-on must have been created with `networkgroup = true`, and the target application must be a member of that networkgroup.

```terraform
resource "clevercloud_otoroshi" "gateway" {
  name         = "my-gateway"
  networkgroup = true
}

resource "clevercloud_nodejs" "app" {
  name               = "my-app"
  min_instance_count = 1
  max_instance_count = 1
  smallest_flavor    = "XS"
  biggest_flavor     = "XS"

  networkgroups = [{
    networkgroup_id = clevercloud_otoroshi.gateway.networkgroup_id
  }]
}

resource "clevercloud_otoroshi_route" "app" {
  otoroshi_id    = clevercloud_otoroshi.gateway.id
  name           = "my-app"
  application_id = clevercloud_nodejs.app.id
  port           = 4242
  path           = "/my-app"
}
```

Domains served by the add-on under `services.clever-cloud.com` are reserved for Otoroshi's own endpoints (console, API, SSO) and cannot carry routes. Any other domain attached to the Otoroshi entrypoint works.
