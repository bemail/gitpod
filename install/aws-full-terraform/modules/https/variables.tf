variable "project_name" {
  type = string
}

variable "gitpod-node-arn" {
  type = string
}

variable "domain" {
  type = string
}

variable "zone_name" {
  type = string
}

variable "region" {
  type = string
}

variable "cluster_name" {
  type = string
}


variable "gitpod_namespace" {
  type = string
}

variable "cert_manager" {
  type = object({
    chart     = string
    email     = string
    namespace = string
  })
}
