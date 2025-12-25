variable "unifi_base_url" {
  description = "The base URL of the UniFi controller (e.g., https://192.168.1.1)"
  type        = string
  default     = ""
}

variable "unifi_api_key" {
  description = "API key for authentication (recommended over username/password)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "unifi_username" {
  description = "Username for authentication (alternative to API key)"
  type        = string
  default     = ""
}

variable "unifi_password" {
  description = "Password for authentication (alternative to API key)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "unifi_site" {
  description = "The UniFi site name"
  type        = string
  default     = "default"
}

variable "unifi_insecure" {
  description = "Skip TLS certificate verification (not recommended for production)"
  type        = bool
  default     = false
}
