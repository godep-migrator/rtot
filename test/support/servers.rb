require 'time'

now = Time.now.utc

rtot_port = rand(13_000..13_010)
Mtbb.register(
  :rtot,
  server_name: 'rtot',
  executable: "#{ENV['GOPATH'].split(/:/).first}/bin/rtot",
  argv: [
    '-a', ":#{rtot_port}",
    '-s', "fizzbuzz#{rtot_port}"
  ],
  port: rtot_port,
  start: now
)
