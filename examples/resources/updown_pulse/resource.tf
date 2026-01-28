# Monitor a daily backup job
resource "updown_pulse" "backup_job" {
  alias   = "nightly-backup"
  period  = 86400 # expect heartbeat every 24 hours
  enabled = true

  recipients = [
    "email:123456789"
  ]
}

# Output the URL to use in your cron job
output "backup_heartbeat_url" {
  value       = updown_pulse.backup_job.pulse_url
  description = "POST to this URL after successful backup"
}

# Monitor a job that runs every 5 minutes
resource "updown_pulse" "frequent_job" {
  alias  = "data-sync"
  period = 300 # expect heartbeat every 5 minutes
}
