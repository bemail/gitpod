variable "project" {
  type    = string
  default = "self-hosted"
}

variable "valueFiles" {
  type    = list(string)
  default = ["./values.yml"]
}

variable "gitpod_namespace" {
  type    = string
  default = "default"
}

variable "region" {
  type = string
}

variable "kubernetes" {
  type = object({
    cluster_name  = string
    version       = string
    node_count    = number
    instance_type = string
  })
  default = {
    cluster_name  = "gitpod-cluster"
    version       = "1.16"
    node_count    = 6
    instance_type = "m4.large"
  }
}

variable "domain" {
  type = string
}

variable "zone_name" {
  type    = string
  default = ""
}

variable "cert_manager" {
  type = object({
    chart     = string
    email     = string
    namespace = string
  })
  default = {
    chart     = ""
    email     = ""
    namespace = ""
  }
}


variable "database" {
  type = object({
    name           = string
    port           = number
    instance_class = string
    engine_version = string
    user_name      = string
    password       = string
  })
  default = {
    name           = "gitpod"
    user_name      = "gitpod"
    password       = ""
    engine_version = "5.7.26"
    port           = 3306
    instance_class = "db.t2.micro"
  }
}


variable "auth_providers" {
  type = list(
    object({
      id            = string
      host          = string
      client_id     = string
      client_secret = string
      settings_url  = string
      callback_url  = string
      protocol      = string
      type          = string
    })
  )
  default = []
}

variable "vpc_name" {
  type    = string
  default = "gitpod-network"
}

variable "includeHTTPS" {
  type    = bool
  default = false
}

variable "includeExternalDB" {
  type    = bool
  default = false
}

variable "bringYourOwnDomain" {
  type    = bool
  default = false
}

variable "includeExternalStorage" {
  type    = bool
  default = false
}
