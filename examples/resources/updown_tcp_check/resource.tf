resource "updown_tcp_check" "my_database" {
  alias   = "PostgreSQL Database"
  url     = "tcp://db.example.com:5432"
  period  = 60
  enabled = true

  recipients = [
    "email:123456789"
  ]
}

resource "updown_tcp_check" "my_tls_service" {
  alias   = "TLS Service"
  url     = "tcps://secure.example.com:443"
  period  = 30
  apdex_t = 1.0
  enabled = true

  disabled_locations = [
    "mia",
  ]
}
