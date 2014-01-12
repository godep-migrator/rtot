require_relative 'test_helper'

describe 'rtot server' do
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
    perform_request(request, port)
  end

  def get(options = {})
    perform_request(
      Net::HTTP::Get.new(options.fetch(:path), 'Rtot-Secret' => secret),
      port
    )
  end

  it 'is pingable without auth' do
    get_request(path: '/', port: port).code.must_equal('200')
  end

  it 'requires auth to create jobs' do
    post_request(
      path: '/', body: 'echo huh',
      port: port
    ).code.must_equal('401')
  end

  it 'creates jobs' do
    post(path: '/', body: 'echo foo').code.must_equal('201')
  end

  it 'returns created job' do
    response = JSON.parse(post(path: '/', body: 'echo huh').body)
    response['jobs'].wont_equal([])
    jobs = response['jobs']
    jobs.length.must_equal(1)
    jobs.first.keys.must_include('create')
  end

  it 'requires auth to get jobs' do
    response = JSON.parse(post(path: '/', body: 'echo hurm').body)
    get_request(
      path: response['jobs'].first['href'],
      port: port
    ).code.must_equal('401')
  end

  it 'returns individual jobs' do
    response = JSON.parse(post(path: '/', body: 'echo hurm ; sleep 1').body)
    get(path: response['jobs'].first['href']).code.must_equal('202')
  end

  it 'requires auth to get all jobs' do
    2.times { |n| post(path: '/', body: "echo #{n} time") }
    get_request(path: '/all', port: port).code.must_equal('401')
  end

  it 'returns all jobs' do
    2.times { |n| post(path: '/', body: "echo #{n} time") }
    response = JSON.parse(get(path: '/all').body)
    response['jobs'].length.must_be(:>, 1)
  end
end
