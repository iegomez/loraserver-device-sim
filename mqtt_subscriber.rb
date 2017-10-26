require 'rubygems'
require 'mqtt'
require 'json'
require 'base64'

# Publish example
#MQTT::Client.connect('test.mosquitto.org') do |c|
#  c.publish('test', 'message')
#end

# Subscribe example
MQTT::Client.connect(
  :host => 'localhost', 
  :port => 1883, 
  :ssl => false,
  :username => 'lora',
  :password => 'lora'
) do |c|
    #c.subscribe("application/1/node/0202020202020202/join")
    #c.subscribe("application/1/node/0202020202020202/rx")
    #c.subscribe("gateway/00800000a00006cd/rx")
    #c.subscribe("gateway/00800000a00006cd/tx")
    c.subscribe("#")

    c.get do |topic, message|
      puts "\n***\n"
      puts "Topic: #{topic}"
      puts "Message: #{message}"
      puts "\n***\n"
    end
end