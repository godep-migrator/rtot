require_relative 'test_helper'

describe 'rtot server' do
  include Mtbb::NetThings

  def port
    @port ||= Mtbb.server(:rtot).port
  end

  def secret
    @secret ||= "fizzbuzz#{port}"
  end

  it 'is pingable' do
    get_request(path: '/', port: port).code.must_equal('200')
  end
end
