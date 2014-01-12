require 'json'
require 'net/http'
require_relative 'support/servers'

include Mtbb::NetThings

def port
  @port ||= Mtbb.server(:rtot).port
end

def secret
  @secret ||= "fizzbuzz#{port}"
end

def post(options = {})
  request = Net::HTTP::Post.new(
    options.fetch(:path), 'Rtot-Secret' => secret
  )
  request.body = options.fetch(:body)
  response = perform_request(request, port)
  { res: response, json: JSON.parse(response.body) }
end

def get(options = {})
  response = perform_request(
    Net::HTTP::Get.new(options.fetch(:path), 'Rtot-Secret' => secret),
    port
  )
  { res: response, json: JSON.parse(response.body) }
end

def delete(options = {})
  response = perform_request(
    Net::HTTP::Delete.new(options.fetch(:path), 'Rtot-Secret' => secret),
    port
  )
  { res: response, json: '' }
end
