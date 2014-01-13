require_relative 'test_helper'

describe 'rtot server' do
  before do
    delete(path: '/jobs')
  end

  it 'is pingable without auth' do
    get_request(path: '/ping', port: port).code.must_equal('200')
  end

  it 'requires auth to create jobs' do
    post_request(
      path: '/jobs', body: 'echo huh',
      port: port
    ).code.must_equal('401')
  end

  it 'creates jobs' do
    post(path: '/jobs', body: 'echo foo')[:res].code.must_equal('201')
  end

  it 'returns created job' do
    response = post(path: '/jobs', body: 'echo huh')[:json]
    response['jobs'].wont_equal([])
    jobs = response['jobs']
    jobs.length.must_equal(1)
    jobs.first.keys.must_include('create')
  end

  it 'requires auth to get jobs' do
    response = post(path: '/jobs', body: 'echo hurm')[:json]
    get_request(
      path: response['jobs'].first['href'],
      port: port
    ).code.must_equal('401')
  end

  it 'returns individual jobs' do
    response = post(path: '/jobs', body: 'echo hurm ; sleep 1')[:json]
    get(path: response['jobs'].first['href'])[:res].code.must_equal('202')
  end

  it 'requires auth to get all jobs' do
    2.times { |n| post(path: '/jobs', body: "echo #{n} time") }
    get_request(path: '/jobs', port: port).code.must_equal('401')
  end

  it 'returns all jobs' do
    2.times { |n| post(path: '/jobs', body: "echo #{n} time") }
    get(path: '/jobs')[:json]['jobs'].length.must_be(:>, 1)
  end

  it 'allows for getting all by state' do
    post(path: '/jobs', body: 'echo fast-ish')
    post(path: '/jobs', body: 'echo slowwwww ; sleep 5')
    sleep 0.5
    get(path: '/jobs?state=running')[:json]['jobs'].length.must_equal(1)
  end

  it 'allows for deleting by state' do
    post(path: '/jobs', body: 'echo fast-ish')
    post(path: '/jobs', body: 'echo slowwwww ; sleep 5')
    sleep 0.5
    delete(path: '/jobs?state=complete')
    get(path: '/jobs')[:json]['jobs'].length.must_equal(1)
  end

  it 'includes "exit" as a string' do
    job = post(path: '/jobs', body: 'exit 1')[:json]['jobs'].first
    sleep 0.1
    exit_str = get(path: job['href'])[:json]['jobs'].first['exit']
    exit_str.must_equal('exit status 1')
  end

  it 'omits "exit" when exit code is 0' do
    job = post(path: '/jobs', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('exit')
  end

  it 'omits "err" when empty' do
    job = post(path: '/jobs', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('err')
  end

  it 'omits "out" when empty' do
    job = post(path: '/jobs', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('out')
  end

  it 'only sends "id", "state", and "href" when "?fields="' do
    job = post(path: '/jobs', body: 'exit 0')[:json]['jobs'].first
    keys = get(path: "#{job['href']}?fields=")[:json]['jobs'].first.keys.sort
	keys.must_equal(%w(href id state))
  end
end
