# The check_id is basically whatever 4 characters you have after https://updown.io/
# It looks like the following regexp : ^https:\/\/updown.io\/([a-z0-9]{4})$

terraform import updown_pulse.my_job <check_id>
